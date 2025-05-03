package main

import (
	"fmt"
	"internal/ai"
	"internal/task"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli/v2"
)

var (
	unselected_style = lipgloss.NewStyle().
				Bold(false).
				Foreground(lipgloss.Color("240")).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("240"))

	selected_style = unselected_style.
			Bold(true).
			Foreground(lipgloss.Color("6")).
			BorderForeground(lipgloss.Color("6"))

	itemStyle         = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("240"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("6"))
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type item struct {
	Name        string
	Description string
	sub_tasks   []item
}

type FocusedState int

const (
	StateTaskBrowser FocusedState = iota
	StateShowForm
)

type model struct {
	formModel         FormInputModel
	confirmModel      FormConfirmModel
	state             FocusedState
	dump              io.Writer // For Debuggging
	list              list.Model
	aiPrompt          string
	aiOutput          string
	aiMsgChannel      chan string
	msgGenerationDone bool
	width             int
	height            int
	confirmquit       bool
	showoverlay       bool
}

// Required by list.Model, ununsed
func (i item) FilterValue() string { return i.Name }

// Returns the initial/default model state
func initialModel() model {
	m := model{}

	var saved_tasks task.TaskFile
	task.ReadTasksFromJson("task_list.json", &saved_tasks)

	var initial_items []list.Item
	for i := range saved_tasks.Tasks {
		task := saved_tasks.Tasks[i]
		initial_items = append(
			initial_items,
			item{
				Name:        task.GetName(),
				Description: task.GetDescription(),
			},
		)
	}

	m.list = list.New(initial_items, itemDelegate{}, 30, 10)
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.SetShowTitle(false)
	m.list.DisableQuitKeybindings()

	m.aiMsgChannel = make(chan string)

	m.formModel.Width = 60

	return m
}

func (m *model) Init() tea.Cmd {
	log.Println("Dr Task Started!")
	m.aiPrompt = "Provide a 2-4 sentence greeting message for today."
	return tea.Batch(m.GenerateAICmd, doTick())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// To enable watching of tea.Msg's using tail (see main)
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case AIStart:
		return m, tea.Batch(m.GenerateAICmd, doTick())

	case TickMsg:
		if !m.msgGenerationDone || len(m.aiMsgChannel) > 0 {
			newContents := <-m.aiMsgChannel
			m.aiOutput += newContents
			return m, doTick()
		}

	case tea.KeyMsg:
		cmd = m.handleKeyInput(msg, nil)
	}

	var subModelCmd tea.Cmd
	switch m.state {
	case StateTaskBrowser:
		m.list, subModelCmd = m.list.Update(msg)

	case StateShowForm:
		switch {
		case m.formModel.Active:
			m.formModel, subModelCmd = m.formModel.Update(msg)
		case m.confirmModel.Active:
			m.confirmModel, subModelCmd = m.confirmModel.Update(msg)
		default:
			m.state = StateTaskBrowser
		}
	}

	return m, tea.Batch(cmd, subModelCmd)
}

func (m *model) handleKeyInput(msg tea.KeyMsg, cmd tea.Cmd) tea.Cmd {
	switch m.state {
	case StateShowForm:
	default:
		switch msg.String() {
		case "ctrl+c", "q":
			m.confirmquit = true
			m.Save()
			return tea.Quit
		case " ":
			return func() tea.Msg {
				m.showoverlay = !m.showoverlay
				return "Overlay changed"
			}
		case "a":
			return func() tea.Msg {
				m.state = StateShowForm
				newItemPos := len(m.list.Items())

				m.formModel.Activate(
					[]string{"Task Name", "Task Description"},
					func(fm *FormInputModel) tea.Msg {
						return m.list.InsertItem(newItemPos, item{Name: fm.inputs[0].Value(), Description: fm.inputs[1].Value()})
					},
					func(fm *FormInputModel) {},
				)
				return "Task model switched!"
			}
		case "d":
			return func() tea.Msg {
				m.state = StateShowForm

				m.confirmModel.Activate(
					func(fm *FormConfirmModel) {
						m.list.RemoveItem(m.list.Index())
					},
					func(m *FormConfirmModel) {},
				)
				return "Confirm model activated!"
			}
		case "p":
			return func() tea.Msg {
				m.state = StateShowForm
				m.formModel.Activate(
					[]string{"AI Prompt"},
					func(fm *FormInputModel) tea.Msg {
						m.aiPrompt = fm.inputs[0].Value()

						return AIStart{}
					},
					func(fm *FormInputModel) {},
				)
				return "AI Request form!"
			}
		}
	}

	return cmd
}

// UI render loop method that is called by bubbletea
func (m *model) View() string {
	if m.confirmquit {
		return ""
	}

	final := renderMainUI(m)

	switch {
	case m.showoverlay:
		text := "This is an overlay!"
		x_padding := m.width / 4
		overlay_container := selected_style.Padding(0, x_padding).Render(text)
		x := (m.width-len(text))/2 - x_padding
		return PlaceOverlay(x, m.height/2-1, overlay_container, final)

	case m.formModel.Active:
		formText := m.formModel.View()
		overlay_container := selected_style.Padding(0, 1).Render(formText)
		x, y := (m.width-m.formModel.Width)/2-1, (m.height-len(m.formModel.inputs))/2-1
		return PlaceOverlay(x, y, overlay_container, final)

	case m.confirmModel.Active:
		formText := m.confirmModel.View()
		overlay_container := selected_style.Padding(0, 1).Render(formText)
		x, y := (m.width-lipgloss.Width(formText))/2-1, (m.height-len(m.formModel.inputs))/2-1
		return PlaceOverlay(x, y, overlay_container, final)
	}

	return final
}

// Renders the main screen of the program
func renderMainUI(m *model) string {
	right_area_width := m.width - m.list.Width() - 4
	top_height := m.height / 5
	bottom_height := 1
	middle_height := m.height - top_height - bottom_height - 8

	ai_bar := selected_style.Width(right_area_width).Height(top_height).Render("Dr. Task\n" + m.aiOutput)
	top := lipgloss.JoinHorizontal(lipgloss.Center, lipgloss.NewStyle().Width(m.width-right_area_width-2).Height(5).Render(""), ai_bar)
	task_list := selected_style.Width(m.list.Width()).Height(middle_height).Render("Task List\n" + m.list.View())

	selected_item, ok := m.list.SelectedItem().(item)
	var task_info string
	if !ok {
		task_info = selected_style.Width(right_area_width).Height(middle_height).Render("Task Description\n" + "Error: Failed to fetch selected item.")
	} else {
		task_info = selected_style.Width(right_area_width).Height(middle_height).Render("Task Description\n" + selected_item.Description)
	}
	middle := lipgloss.JoinHorizontal(lipgloss.Center, task_list, task_info)

	footer := selected_style.Width(m.width - 2).Height(bottom_height).Render("Controls: a - Add task | d - Delete Item | p - Prompt Dr. Task | q - Quit")

	full := lipgloss.JoinVertical(lipgloss.Center, top, middle, footer)
	final := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, full)

	return final
}

// Helper functions

type AIStart struct{}

type AIFinished struct {
	content string
}

// tea.Cmd for sending a request to the locally running AI to generate a new message
func (m *model) GenerateAICmd() tea.Msg {
	m.msgGenerationDone = false
	m.aiOutput = ""
	ai.AiRequestStreamedChannel(
		`You are 'Dr. Task', an assistant whose goal is to provide an overview and advice for a user of a task management program.
		You must:
		- Use a formal, working tone of voice.
		- Act more like an assistant than a boss.
		- Don't be excessively enthuseastic
		`,
		m.aiPrompt,
		m.aiMsgChannel,
	)
	m.msgGenerationDone = true

	return AIFinished{content: m.aiOutput}
}

// Save the current state to a json file
func (m model) Save() {
	var saved_tasks task.TaskFile
	list_items := m.list.Items()

	for i := range list_items {
		it := list_items[i].(item)
		saved_tasks.Tasks = append(saved_tasks.Tasks, task.Task{
			Name:        it.Name,
			Description: it.Description,
			Created:     time.Now(),
			Due:         time.Now(),
		})
	}

	task.SaveTasksToJson(saved_tasks, "task_list.json")
}

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func main() {
	// Debugging
	// - set DEBUG (`export DEBUG=1`) and `tail -f messages.log` in another terminal
	var dump *os.File
	if _, ok := os.LookupEnv("DEBUG"); ok {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(1)
		}
	}

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	app := &cli.App{
		Name:  "drtask",
		Usage: "A simple terminal based task organiser with AI support.",
		Action: func(*cli.Context) error {

			initialModel := initialModel()
			initialModel.dump = dump
			p := tea.NewProgram(&initialModel)

			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Task struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	Due         time.Time `json:"due"`
}

type TaskFile struct {
	Tasks []Task `json:"tasks"`
}

func (t Task) GetName() string        { return t.Name }
func (t Task) GetDescription() string { return t.Description }
func (t Task) GetCreated() time.Time  { return t.Created }
func (t Task) GetDue() time.Time      { return t.Due }

func (t Task) PrettyString() string {
	return fmt.Sprintf(
		"Task: %s    Description: %s\nCreated: %s    Due: %s",
		t.GetName(),
		t.GetDescription(),
		t.GetCreated().String(),
		t.GetDue().String(),
	)
}

func SaveTasksToJson(task_file TaskFile, filename string) error {
	jsonData, err := json.MarshalIndent(task_file, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func ReadTasksFromJson(filename string, v *TaskFile) error {
	file_as_bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error when reading the file: %v", err)
	}

	if err := json.Unmarshal(file_as_bytes, v); err != nil {
		return err
	}

	return nil
}

func AddTask(tasks *[]Task, task Task) {
	*tasks = append(*tasks, task)
}

func StringFromTime(t time.Time) string {
	time_format := "2006-01-02T15:04:05"

	return t.Format(time_format)
}

func TimeFromString(time_string string) time.Time {
	time_format := "2006-01-02T15:04:05"

	time_value, err := time.Parse(time_format, time_string)

	if err != nil {
		log.Fatalf("Failed to format the time: %s", time_string)
	}

	return time_value
}

func JsonStandardTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func main() {
	var task_file TaskFile
	if err := ReadTasksFromJson("temp.json", &task_file); err != nil {
		log.Fatalf("Error while trying to read the task json file: %v", err)
	}

	SaveTasksToJson(task_file, "temp.json")
}

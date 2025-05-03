package task

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestStringFromTime(t *testing.T) {
	want := "2020-05-03T15:25:39"
	input := time.Date(2020, 05, 03, 15, 25, 39, 0, time.UTC)

	output := StringFromTime(input)

	if output != want {
		t.Errorf(`StringFromTime(time.Date(2020, 05, 03, 15, 25, 39, 0, time.UTC)) = %q, want %q`, output, want)
	}
}

func TestTimeFromString(t *testing.T) {
	want := time.Date(2020, 05, 03, 15, 25, 39, 0, time.UTC)
	input := "2020-05-03T15:25:39"

	output := TimeFromString(input)

	if output != want {
		t.Errorf(`TimeFromString("2020-05-03T15:25:39") = %q, want %q`, output, want)
	}
}

func TestJsonStandardTime(t *testing.T) {
	want := "2020-05-03T15:25:39.001Z"
	input := time.Date(2020, 05, 03, 15, 25, 39, 1000000, time.UTC)

	output := JsonStandardTime(input)

	if output != want {
		t.Errorf(`JsonStandardTime(time.Date(2020, 05, 03, 15, 25, 39, 1, time.UTC)) = %q, want %q`, output, want)
	}
}

func TestReadTasksFromJson(t *testing.T) {
	tempFile, err := os.CreateTemp("", "tasks.json")
	if err != nil {
		t.Fatalf("Error creating temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	task_1_created := TimeFromString("2006-01-02T15:04:05")
	task_2_created := TimeFromString("2007-02-03T16:05:06")

	task_1 := Task{
		Name:        "Test Task 1",
		Description: "Test Task 1 description",
		Created:     task_1_created,
		Due:         task_1_created.Add(time.Hour * 24),
	}
	task_2 := Task{
		Name:        "Test Task 2",
		Description: "Test Task 2 description",
		Created:     task_2_created,
		Due:         task_2_created.Add(time.Hour),
	}

	tasksData := fmt.Appendf(nil, `{
		"tasks": [
			{
				"name": "%s",
				"description": "%s",
				"created": "%s",
				"due": "%s"
			},
			{
				"name": "%s",
				"description": "%s",
				"created": "%s",
				"due": "%s"
			}
  ]
}`,
		task_1.Name, task_1.Description, JsonStandardTime(task_1.Created), JsonStandardTime(task_1.Due),
		task_2.Name, task_2.Description, JsonStandardTime(task_2.Created), JsonStandardTime(task_2.Due),
	)

	if _, err := tempFile.Write(tasksData); err != nil {
		t.Fatalf("Error writing to temporary file: %v", err)
	}
	tempFile.Close()

	var task_file TaskFile
	err = readTasksFromJson(tempFile.Name(), &task_file)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	tasks := task_file.Tasks

	if len(tasks) != 2 || tasks[0] != task_1 || tasks[1] != task_2 {
		t.Errorf("tasks were not read correctly. got: %v", tasks)
	}
}

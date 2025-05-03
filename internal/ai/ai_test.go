package ai

import (
	"fmt"
	"testing"
)

func TestStringFromTime(t *testing.T) {
	fmt.Println(AiResuestStreamed("Be creative", "Write a haiku"))
	// want := "2020-05-03T15:25:39"
	// input := time.Date(2020, 05, 03, 15, 25, 39, 0, time.UTC)
	//
	// output := StringFromTime(input)

	// if output != want {
	// 	t.Errorf(`StringFromTime(time.Date(2020, 05, 03, 15, 25, 39, 0, time.UTC)) = %q, want %q`, output, want)
	// }
}

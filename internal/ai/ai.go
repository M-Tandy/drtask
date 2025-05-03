package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const api_endpoint = "http://localhost:1234/v1/chat/completions"

func AiRequest(system_prompt string, user_prompt string) string {
	client := resty.New()

	request_body := map[string]any{
		"model": "qwen2.5-coder-14b-instruct",
		"messages": []any{
			map[string]any{"role": "system", "content": system_prompt},
			map[string]any{"role": "user", "content": user_prompt},
		},
		"temperature": 0.7,
		"max_tokens":  -1,
		"stream":      false,
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request_body).Post(api_endpoint)

	if err != nil {
		log.Fatalf("Error sending the request: %v", err)
	}

	body := response.Body()

	var body_as_string map[string]any
	err = json.Unmarshal(body, &body_as_string)

	if err != nil {
		log.Fatalf("Error while tring to decode json response to string: %v\nGot response: %v", err, body)
		return ""
	}

	content := body_as_string["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)
	return content
}

func AiResuestStreamed(system_prompt string, user_prompt string, response_body *string) {
	client := http.Client{}

	request_body := map[string]any{
		"model": "qwen2.5-coder-14b-instruct",
		"messages": []any{
			map[string]any{"role": "system", "content": system_prompt},
			map[string]any{"role": "user", "content": user_prompt},
		},
		"temperature": 0.7,
		"max_tokens":  -1,
		"stream":      true,
	}
	m_request_body, err := json.Marshal(request_body)
	if err != nil {
		log.Fatalf("Error while marshalling request body: %v", err)
	}
	reader := bytes.NewReader(m_request_body)

	req, err := http.NewRequest("POST", api_endpoint, reader)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error while creating response: %v", err)
	}
	defer resp.Body.Close()

	var v map[string]any
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()

		// ChatGPT api responses when streaming start with 'data: '. The final response is 'data: [DONE]'
		if len(line) == 0 {
			continue
		}
		if line == "data: [DONE]" {
			break
		}
		if line[:7] == "data: {" {
			err := json.Unmarshal([]byte(line[6:]), &v)
			if err != nil {
				fmt.Println("Error unmarshalling JSON:", err)
				continue
			}
			delta := v["choices"].([]any)[0].(map[string]any)["delta"].(map[string]any)["content"]
			if delta != nil {
				*response_body = *response_body + delta.(string)
			}
		}
	}
}


func AiRequestStreamedChannel(system_prompt string, user_prompt string, response_body chan<- string) {
	client := http.Client{}
	request_prompt := user_prompt

	request_body := map[string]any{
		"model": "qwen2.5-coder-14b-instruct",
		"messages": []any{
			map[string]any{"role": "system", "content": system_prompt},
			map[string]any{"role": "user", "content": request_prompt},
		},
		"temperature": 0.7,
		"max_tokens":  -1,
		"stream":      true,
	}
	m_request_body, err := json.Marshal(request_body)
	if err != nil {
		log.Fatalf("Error while marshalling request body: %v", err)
	}
	reader := bytes.NewReader(m_request_body)

	req, err := http.NewRequest("POST", api_endpoint, reader)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error while creating response: %v", err)
	}
	defer resp.Body.Close()

	var v map[string]any
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()

		// ChatGPT api responses when streaming start with 'data: '. The final response is 'data: [DONE]'
		if len(line) == 0 {
			continue
		}
		if line == "data: [DONE]" {
			log.Printf("Completed AI response for: %s", request_prompt)
			break
		}
		if line[:7] == "data: {" {
			err := json.Unmarshal([]byte(line[6:]), &v)
			if err != nil {
				log.Printf("Error unmarshalling JSON:\n", err)
				continue
			}
			delta := v["choices"].([]any)[0].(map[string]any)["delta"].(map[string]any)["content"]
			if delta != nil {
				// *response_body = *response_body + delta.(string)
				response_body <- delta.(string)
			}
		}
	}
}

func dateSubstitute(original string) string {
	current_date := time.Now().Local()
	current_day := current_date.Weekday()

	return strings.ReplaceAll(strings.ReplaceAll(original, "{%date%}", current_date.Format("January 02 2006")), "{%weekday%}", current_day.String())
}

// func main() {
// 	data, err := os.ReadFile("./greetings.txt")
//
// 	if err != nil {
// 		fmt.Printf("Error when reading the file: %v", err)
// 	}
//
// 	split_strings := strings.Split(string(data), "\n")
//
// 	for i := range split_strings {
// 		fmt.Println(date_substitute(split_strings[i]))
// 	}
// }

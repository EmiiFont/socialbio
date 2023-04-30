package main

import (
	"context"
	"fmt"
	"net/http"
	"text/template"

	"github.com/sashabaranov/go-openai"
)

type SocialBio struct {
	Bio      string
	Style    string
	Emojis   bool
	Language string
}

func hello(w http.ResponseWriter, req *http.Request) {
	t := template.New("index.html")
	t, err := t.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}
	p := SocialBio{Bio: "John"}
	t.Execute(w, p)
}

//build submit handler to handle post request of a form
func submit(w http.ResponseWriter, req *http.Request) {
	//check if the request is a post request
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	//get the form value
	req.ParseForm()
	//get the username
	bio := req.FormValue("bio")
	style := req.FormValue("style")
	emojis := req.FormValue("emojis")
	language := req.FormValue("language")

	//create a person object
	p := SocialBio{Bio: bio, Style: style, Emojis: emojis == "on", Language: language}

	client := openai.NewClient("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")
	prompt := fmt.Sprintf("Generate a instagram %s bio using %s! in %s", p.Style, p.Bio, p.Language)
	if p.Emojis {
		prompt += "and include emojis"
	}

	resp, chatErr := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if chatErr != nil {
		panic(chatErr)
	}

	fmt.Printf(resp.Choices[0].Message.Content)

	// //create a template
	// t := template.New("index.html")
	// //parse the template
	// t, err := t.ParseFiles("templates/index.html")
	// if err != nil {
	// 	panic(err)
	// }
	// //execute the template
	// t.Execute(w, p)
	w.Write([]byte(resp.Choices[0].Message.Content))
}

func main() {

	http.HandleFunc("/", hello)
	// handle post request to /submit
	http.HandleFunc("/submit", submit)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":4000", nil)
}

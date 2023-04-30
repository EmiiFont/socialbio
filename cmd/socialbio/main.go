package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
	files := []string{
		"templates/base.tmpl",
		"templates/main.tmpl",
		"templates/footer.tmpl",
		"templates/nav.tmpl",
		"templates/content.tmpl",
	}

	t := template.New("index.html")
	t, err := t.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	err = t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		panic(err)
	}
	//t.Execute(w, p)
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

	openaiKey := os.Getenv("OPENAI_KEY")
	client := openai.NewClient(openaiKey)
	emojisPrompt := "don't include"
	if p.Emojis {
		emojisPrompt = "include"
	}

	prompt := fmt.Sprintf("Generate 1 instagram %s biography with no hashtags and clearly labeled and make sure each generated biography is less than 160 characters using %s! in %s and %s emojis", p.Style, p.Bio, p.Language, emojisPrompt)

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
	w.Write([]byte(resp.Choices[0].Message.Content))
}

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets/"))))
	http.HandleFunc("/", hello)
	// handle post request to /submit
	http.HandleFunc("/submit", submit)

	fmt.Println("✅ Server up and running on port: " + port)
	http.ListenAndServe(":"+port, nil)
}

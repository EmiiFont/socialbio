package main

import (
	"context"
	"embed"
	"fmt"
	"go/build"
	"io/fs"
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

const (
	layoutsDir   = "templates"
	templatesDir = "templates"
	extension    = "/*.tmpl"
)

var (
	//go:embed templates/*
	files     embed.FS
	templates map[string]*template.Template
	//go:embed assets/*
	assets embed.FS
)

func LoadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		return err
	}
	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name(), layoutsDir+extension)
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt
	}
	return nil
}

func hello(w http.ResponseWriter, req *http.Request) {
	t, ok := templates["base.tmpl"]
	if !ok {
		fmt.Println(templates)
		panic("template not found")
	}
	err := t.ExecuteTemplate(w, "base", nil)
	if err != nil {
		panic(err)
	}
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

	prompt := fmt.Sprintf("Generate an twitter %s bio with no hashtags using %s! and in %s language, please %s emojis", p.Style, p.Bio, p.Language, emojisPrompt)

	resp, chatErr := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: 200,
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
	err := LoadTemplates()
	if err != nil {
		panic(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))
	http.HandleFunc("/", hello)
	// handle post request to /submit
	http.HandleFunc("/submit", submit)

	fmt.Println(os.Getwd())
	fmt.Println(build.Default.GOPATH)
	fmt.Println("✅ Server up and running on port: " + port)
	http.ListenAndServe(":"+port, nil)
}

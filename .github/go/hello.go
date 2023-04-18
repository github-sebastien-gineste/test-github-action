package main

import (
	"fmt"
	"os"
	"context"

    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
)

func main() {
	// Récupérer le token d'authentification depuis les secrets de Github
	token := os.Getenv("GITHUB_TOKEN")

	// Créer un client Github avec l'authentification
	ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(oauth2.NoContext, ts)
    client := github.NewClient(tc)

	 // Récupérer les informations de la Pull Request
	 prNumber := os.Getenv("PR_NUMBER")
	 owner := os.Getenv("OWNER")
	 repo := os.Getenv("REPO")
	 pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, prNumber)
	 if err != nil {
		 fmt.Println(err)
		 return
	 }

	// Ajouter un commentaire à la Pull Request
	comment := &github.IssueComment{
		Body: github.String("Coucou"),
    }
    _, _, err = client.Issues.CreateComment(context.Background(), owner, repo, pr.GetNumber(), comment)
    if err != nil {
        fmt.Println(err)
        return
    }

	fmt.Println("Commentaire ajouté avec succès")
}


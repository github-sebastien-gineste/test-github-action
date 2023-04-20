package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	prNumberStr := os.Getenv("PR_NUMBER")
	prNumber, err := strconv.Atoi(prNumberStr)
	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, prNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Récupérer les fichiers modifiés dans la Pull Request
	files, _, err := client.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Afficher les fichiers modifiés
	prBody := os.Getenv("PR_BODY")

	fmt.Println("Fichiers modifiés dans la Pull Request :")
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, *file.Filename)
	}
	filesStr := strings.Join(filenames, "\n")

	// Ajouter un commentaire à la Pull Request
	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf("Coucou ! Voici la liste des fichiers modifiés dans cette Pull Request : \n\n%s \n %s", filesStr, prBody)),
	}
	_, _, err = client.Issues.CreateComment(context.Background(), owner, repo, pr.GetNumber(), comment)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Commentaire ajouté avec succès")
}

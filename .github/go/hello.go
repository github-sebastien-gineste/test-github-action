package main

import (
	"bufio"
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
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")

	// Créer un client Github avec l'authentification
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Récupérer les informations de la Pull Request
	prNumberStr := os.Getenv("PR_NUMBER")
	prNumber, err := strconv.Atoi(prNumberStr)
	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Titre de la Pull Request : ", pr.GetTitle())
	fmt.Println("Body de la Pull Request : ", pr.GetBody())
	fmt.Println("Is_editable : ", pr.GetMaintainerCanModify())

	// Récupérer les fichiers modifiés dans la Pull Request
	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
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
		Body: github.String(fmt.Sprintf("Coucou ! Voici la liste des fichiers modifiés dans cette Pull Request : \n\n%s \nBody : \n %s", filesStr, prBody)),
	}

	_, _, err = client.Issues.CreateComment(ctx, owner, repo, pr.GetNumber(), comment)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Lire le contenu du fichier check.md
	file, err := os.Open("../template/proto_checklist.md")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	//body := strings.Join(bodyLines, "\n")

	// Mettre à jour le corps de la Pull Request avec  le contenu du fichier  check.md
	pr.MaintainerCanModify = github.Bool(true)
	pr.Body = github.String("body")

	fmt.Println("Titre de la Pull Request : ", pr.GetTitle())
	fmt.Println("Body de la Pull Request : ", pr.GetBody())
	fmt.Println("Is_editable : ", pr.GetMaintainerCanModify())

	u := fmt.Sprintf("repos/%v/%v/pulls/%d", owner, repo, pr.GetNumber())

	println(u)

	_, _, err = client.PullRequests.Edit(ctx, owner, repo, pr.GetNumber(), pr)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Commentaire ajouté avec succès")
}

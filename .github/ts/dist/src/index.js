"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const github_1 = require("@actions/github");
const core_1 = require("@actions/core");
const githubToken = process.env.GITHUB_TOKEN;
const github = (0, github_1.getOctokit)(githubToken);
const context = github_1.context;
github_action();
async function github_action() {
    let message = "coucou from TS  \n\n";
    const files_names = await getDiff();
    files_names.map((file) => {
        console.log(file.filename);
        message += ` - ${file.filename}  \n`;
    });
    (0, core_1.notice)('Something happened that you might want to know about.');
    message += "Body : " + process.env.PR_BODY;
    // Déclenche l'évenement que lors d'un push vers la PR 
    // Get the last Dif of the PR
    // Compare the files names changed
    // If the file name is new in the checklist
    // Add its checklist
    // Else if a file name disapeer
    // Remove its checklist
    // Else
    // Do nothing
    // Split the readme in specific checklist 
    await github.rest.pulls.update({
        owner: context.repo.owner,
        repo: context.repo.repo,
        pull_number: context.payload.pull_request.number,
        body: "coucou from TS"
    });
    await github.rest.issues.createComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: context.issue.number,
        body: message
    });
}
async function getDiff() {
    if (githubToken && context.payload.pull_request) {
        const result = await github.rest.repos.compareCommits({
            owner: context.repo.owner,
            repo: context.repo.repo,
            base: context.payload.pull_request.base.sha,
            head: context.payload.pull_request.head.sha,
            per_page: 100
        });
        return result.data.files || [];
    }
    return [];
}

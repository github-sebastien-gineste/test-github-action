"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const github_1 = require("@actions/github");
async function github_action() {
    const githubToken = process.env.GITHUB_TOKEN;
    if (!githubToken) {
        return;
    }
    const pullRequestNumber = github_1.context.payload.pull_request ? github_1.context.payload.pull_request.number : -1;
    const octokit = (0, github_1.getOctokit)(githubToken);
    await octokit.rest.issues.createComment({
        ...github_1.context.repo,
        issue_number: pullRequestNumber,
        body: "Coucou from TS"
    });
}
github_action();

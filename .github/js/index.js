module.exports = async ({github, context, core})=> {
    const { Octokit } = require("@octokit/rest");
    const message = "coucou";

    const octokit = new Octokit({
        auth: process.env.GITHUB_TOKEN,
    });

    await octokit.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: message
    })
}
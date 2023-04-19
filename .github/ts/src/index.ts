// --------  diff --------
import { context as ctx, getOctokit } from "@actions/github";
import { getInput } from "@actions/core"; // core 

const githubToken = process.env.GITHUB_TOKEN!;
const github = getOctokit(githubToken);
const context = ctx;
github_action()
// --------  diff --------

async function github_action() {

    // ----- Same -----  
    const message = "coucou from TS";

    await github.rest.issues.createComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: context.issue.number,
        body: message
    });
    // ----- Same -----  
}


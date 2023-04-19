import { context as ctx, getOctokit } from "@actions/github";
import { getInput } from '@actions/core';

const githubToken = process.env.GITHUB_TOKEN!;
const github = getOctokit(githubToken);
const context = ctx;
github_action()

async function github_action() {

    let message = "coucou from TS  \n\n";


    const files_names = await getDiff();
    files_names.map((file) => {
        console.log(file.filename);
        message += ` - ${file.filename}  \n`;
    });

    const prBody = getInput('pr_body');
    message += "Body : " + prBody;


    await github.rest.issues.createComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: context.issue.number,
        body: message
    });
}

async function getDiff(){
    if(githubToken && context.payload.pull_request){
        const result = await github.rest.repos.compareCommits({
            owner: context.repo.owner,
            repo: context.repo.repo,
            base: context.payload.pull_request.base.sha,
            head: context.payload.pull_request.head.sha,
            per_page: 100
        })

        return result.data.files || [];
    }
    return [];
}
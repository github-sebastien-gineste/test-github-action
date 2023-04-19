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

    const files_names = await getDiff();
    files_names.map((file) => {
        console.log(file.filename);
    });

    // ----- Same -----  
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
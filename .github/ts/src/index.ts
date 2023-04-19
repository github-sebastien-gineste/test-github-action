import { context, getOctokit } from "@actions/github";

async function github_action() {
  const githubToken = process.env.GITHUB_TOKEN;
  if(!githubToken) {return}

  const pullRequestNumber : number = context.payload.pull_request? context.payload.pull_request.number : -1;

  const octokit = getOctokit(githubToken);
  await octokit.rest.issues.createComment({
      ...context.repo,
      issue_number: pullRequestNumber,
      body: "Coucou"
    });
}

github_action()
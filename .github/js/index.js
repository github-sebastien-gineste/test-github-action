module.exports = async ({github, context, core})=> {
    const { owner, repo } = context.repo;
    const number = context.payload.pull_request.number;
    const message = "coucou";
    const octokit = new github.GitHub(process.env.GITHUB_TOKEN);
    await octokit.issues.createComment({
        owner,
        repo,
        issue_number: number,
        body: message
    });
}
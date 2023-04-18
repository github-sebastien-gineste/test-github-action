module.exports = async ({github, context, core})=> {
    const message = "coucou";
 
    await github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: message
    })
}
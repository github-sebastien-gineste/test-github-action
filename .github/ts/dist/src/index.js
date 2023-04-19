module.exports = async ({github, context, core})=> {  
    const message = "coucou from TS";
    await github.rest.issues.createComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: context.issue.number,
        body: message
    });
    }
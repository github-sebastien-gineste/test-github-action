module.exports = async ({github, context, core})=> {
    const execa = require('execa')

    const { stdout } = await execa('echo', ['hello', 'world'])

    console.log(stdout)
    
    const message = "coucou from JS";
 
    await github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: message
    })
}
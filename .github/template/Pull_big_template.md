_**All unnecessary sections of this template should be removed**<br>
So that it represents your changes as closely as possible and facilitates review by the different teams._

# Description

_Please describe your changes in detail and provide related JIRA ticket(s)._

# Reminders

- 👷‍♂️ **Open your PR as a draft** to not notify the platform team until the PR is ready
- **Don't ping the platform team** for a code review on [`#innov-platform-eng`](https://teads.slack.com/archives/CD3GJ2MU5), unless:
  - you have no answer after a reasonable delay (24 hours),
  - your request is urgent.
- Feel free to ask for help on [`#innov-platform-eng`](https://teads.slack.com/archives/CD3GJ2MU5)
- ⚠️ Screenshots included in GitHub **are not allowed** as they are [publicly accessible without authentication](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files). You can provide screenshots through the associated JIRA card.

# Checklist for a proto PR

- [ ] My RPC is in the right domain
- [ ] My RPC's name respects [the naming conventions](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToNameMyRpc.md)
- [ ] My RPC's fields are enriched with [validation rules](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/ValidationRules.md)
- [ ] If it is a command, my RPC declares all the [references](https://github.com/ebuzzing/service-api-domains#reference) necessary for the audit
- [ ] My RPC is [restricted](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToRestrictAnRpcToSpecificClients.md) at least to a specific client
- [ ] My RPC must have a [SecurityContext](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/IdentificationAndAuthentication.md#identification1) except in the _rare cases_ where there is no real user involved thus becoming an [Anonymously RPC](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/IdentificationAndAuthentication.md#:~:text=the%20rpc%20should%20be%20suffixed%20with%20anonymously)
- [ ] My proto describe all the reasons that could make my RPC fail - in case they are complex or they are many - by leveraging [rich errors descriptions](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToUseRichErrorMessages.md) or else I assume errors will flow to the client as a simple verbatim string. 
- [ ] New enums, if any, are embedded in messages except if they are reused somewhere else. 

# Checklist for an implementation PR
- [ ] I have carefully defined the authorization rules of my RPC, aka handler's `isAuthorized`. These authorization rules prevent data breach: my RPC can not alter neither give access to data owned by someone else (user or organisation).

# Checklist for a change in development configuration `[your-domain].conf`
- [ ] I have reflected my additions/changes in production's configuration `api-domains.conf`

# Checklist for a change in production's configuration `api-domains.conf`
- [ ] I have launched a deploy-PR (Please add some logs to validate it).
  ⚠️ If the PR contains a SQL migration, deploy-PR **is not possible** since the migration won't be applied.

# Checklist for a PR containing SQL migrations
- [ ] The migrations are [only applied on api-domains MySQL cluster](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/ADR/OnlyApplyMigrationsToApiDomainsCluster.md)
- [ ] I have tested my migration in a sandbox (SQL migrations cannot be tested in a deploy-PR to avoid modifying production's schemas)
- [ ] I will [deploy api-domains-migrations](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToDeployApiDomainsMigrations.md) first then `api-domains`

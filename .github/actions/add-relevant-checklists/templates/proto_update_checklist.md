# Checklist for an edition proto PR 

- [ ] My RPC's fields are enriched with [validation rules](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/ValidationRules.md)
- [ ] If it is a command, my RPC declares all the [references](https://github.com/ebuzzing/service-api-domains#reference) necessary for the audit
- [ ] New enums, if any, are embedded in messages except if they are reused somewhere else. 

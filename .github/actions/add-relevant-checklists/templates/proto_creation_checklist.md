# Checklist for an addition proto PR

- [ ] My RPC is in the right domain
- [ ] My RPC's name respects [the naming conventions](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToNameMyRpc.md)
- [ ] My RPC's fields are enriched with [validation rules](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/ValidationRules.md)
- [ ] If it is a command, my RPC declares all the [references](https://github.com/ebuzzing/service-api-domains#reference) necessary for the audit
- [ ] My RPC is [restricted](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToRestrictAnRpcToSpecificClients.md) at least to a specific client
- [ ] My RPC must have a [SecurityContext](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/IdentificationAndAuthentication.md#identification1) except in the _rare cases_ where there is no real user involved thus becoming an [Anonymously RPC](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/Explanation/IdentificationAndAuthentication.md#:~:text=the%20rpc%20should%20be%20suffixed%20with%20anonymously)
- [ ] My proto describe all the reasons that could make my RPC fail - in case they are complex or they are many - by leveraging [rich errors descriptions](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToUseRichErrorMessages.md) or else I assume errors will flow to the client as a simple verbatim string. 
- [ ] New enums, if any, are embedded in messages except if they are reused somewhere else. 

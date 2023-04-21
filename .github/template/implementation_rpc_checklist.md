# Checklist for an implementation PR
- [ ] I have carefully defined the authorization rules of my RPC, aka handler's `isAuthorized`. These authorization rules prevent data breach: my RPC can not alter neither give access to data owned by someone else (user or organisation).

---
name: Story Kick Off and DOD
about: Detail information and updates on a story and providing definition of done
title: SMP-
labels: story
assignees: ''

---

https://saltpayco.atlassian.net/browse/SMP-


# Kick-off

## What?

What does the end result look like? Try to visualise yourself doing a functional demo. What will you show?

Open an architecture diagram to show this better, if that helps.

e.g. 
Putting a payment instruction on the ISB unprocessed queue will send that to ISB and we'll be able to query the status of that payment



### Example criteria

Given a valid payment instruction
When it is sent to the ISB service
And we query the status of the payment instruction
Then the status should be processed


## How?

- How are you going to proceed? 
- What's the smallest change _that can be integrated_, and that gets you towards your final goal? Think about "steel thread" through the system. Could you maybe hard-code some values first to move forward?
- What's the next change after that? And after that? 
- Can you see a path made of small steps that get us to our destination? 
- Do you need any feature flags?

Write down your high-level path to getting this work done here. If you can't, maybe have a conversation with someone about how to proceed.

e.g.
1. Write an acceptance test for the whole scenario
2. Define the interface for an ISB adapter that will provided to the makePayment usecase
3. Implement the ISB adapter as an HTTP client (with integration tests) - happy path
4. Handle different kinds of HTTP errors
5. Write unit tests for the makePayment use case calling a mock adapter and handling various return values
6. Plug in the real ISB adapter

Things to note:
- API key / credentials to be added to secrets
- Re-use the HTTP client from BC; create a tech debt to extract it out

### Time estimation

Given the high-level path, how long (days) will each task take. Using this can you estimate the time it will take to complete this story. By what date should you expect this story to be completed by. 

e.g. 
3 day estimate finished by 25-05-2022

# Definition of Done

## Testing 
- [ ] Unit
- [ ] Integration  
- [ ] Acceptance
- [ ] Black box

## Operability 

- [ ] Metrics
    - [ ] Any times to be measured? http response times, sql query times?
    - [ ] Any error metrics? This will allow alerting.  
- [ ] Logging
    - [ ] Excessive logging?
    - [ ] Any chance of secrets being logged?
- [ ] Alerts 
    - [ ] e.g. on errors


## Documentation 

- [ ] Do you need an ADR?
- [ ] Do you need to update the system context diagram?
- [ ] Do you need to update the code architecture diagram?
- [ ] Have you added a new API? Document an example request
- [ ] Have you changed the way the system is built or requires developers to do extra setup? Document in the README.md
- [ ] Have you considered how this change may affect on-call? If so, update the runbook.

## Delivering value

- [ ] Is it live?
- [ ] Have you demonstrated the change to whoever needs it?

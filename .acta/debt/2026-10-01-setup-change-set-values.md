---
id: DBT-0064
hash: o9ou1hj
parent: plans/2026-10-01-setup-change-set-values
---
# Review NOTEs: Setup offers to change values that are already set Implementation Plan

- [ ] (medium) acta config show marks only plan_depth as (default); chat_language and style print a plain value when unset, so setup cannot see an unset voice on a partial config and only offers it through the change-anything question.
- [ ] (low) acta config show says nothing about the acta block, so setup cannot tell whether that part is set yet.
- [ ] (low) The Split subagent models rule (ask only while there is no subagent_models line) clashes with a later run where the user names subagent models to change.
- [ ] (low) The setup heading ## First run now covers later runs too; a name like ## Run would read better.
- [ ] (low) In setup step 2 the named-part bullet sits after the Any other run clause; putting the two branches on their own lines would stop a literal reader applying it to a first run.

# Unslop

Use this list while you review a spec you wrote. Each pattern is a reason to
look at a sentence again, not an automatic error. Change the sentence only when
a plainer one says the same thing. Rule numbers are fixed ids that other
skills may cite; a removed rule leaves a gap.

## What to keep

- Keep the meaning and the intended tone.
- Keep the requirements, constraints, uncertainty, evidence, risks and approval gates. Each one stays as strong as it was.
- Never change code, commands, flags, identifiers, literal errors or direct quotes.
- Never invent a source or a number. When a claim has no source, cut it or say it is unknown.
- The project's own rules still decide what a spec must contain.
- Labels and short status text may stay short. Use whole sentences when a fragment makes the reader guess.

## Content

3. Empty -ing phrases ("highlighting...", "ensuring...", "reflecting..."). Delete them, or say the real fact.
5. Vague sources ("Experts believe", "Some argue"). Name the source or delete the claim.

## Language

7. AI words: additionally, crucial, delve, enhance, fostering, garner, interplay, intricate, pivotal, showcase, testament, underscore, vibrant, abstract "landscape" or "tapestry". Use plain words.
8. Fancy ways to say "is": "serves as", "stands as", "boasts", "features". Say "is" or "has".
9. "Not just X, but Y." State the point directly.
10. Forced groups of three. Use the number of items there really are.
11. Synonym cycling. Pick one name for a thing and repeat it.
12. False ranges ("from X to Y" with no real scale). List the items.

## Style

13. Dashes used as pauses. End the sentence or use a comma.
14. Colons as mid-sentence connectors. A colon is fine before a list or an example.
15. Bold for emphasis. Bold only labels a reader scans for.
16. Lists with a bold inline header on every item. Write the item as a plain sentence unless the label helps the reader find it.
17. Title Case headings. Use sentence case.
18. Decorative emoji. Remove them.
19. Curly quotes. Use straight quotes.

## Chat leftovers

20. Chatbot phrases ("I hope this helps", "Certainly!"). Remove.
22. Flattery ("Great question!"). Remove.

## Filler

23. Filler phrases. "In order to" becomes "To". "It is important to note that" goes.
24. Stacked hedges. "Could potentially possibly" becomes "may". Keep one hedge when the doubt is real.
25. Generic endings ("The future looks bright"). State the plan or the fact.

## Jargon

26. Abstract metaphor nouns: substrate, wedge, vector, locus, nexus, surface, bedrock, scaffolding, paradigm, north star, flywheel, endgame. Use the concrete word.

## Plain speech

27. Say what a thing does, not how it feels. Name the mechanism or the number. If you cannot restate a sentence as an instruction, a fact or a number, cut it. If the sentence could appear unchanged in another project's docs, make it specific or cut it.
28. Shorten or split dense sentences. If the reader has to backtrack to parse a sentence, break it in two or drop clauses. One idea per sentence.
29. Passive voice hides the actor. "The file is parsed by the loader" becomes "the loader parses the file". Passive is fine when the actor is unknown or does not matter.
30. Adverbs propping up weak verbs. "Significantly improves" becomes the measured change.
31. Fancy words. "Utilize" and "leverage" become "use", "facilitate" becomes "help", "in the event that" becomes "if".
32. Mannered prose: aphorisms, personified code, figurative verbs ("rides along"). Say the literal thing.
33. Over-compression: dropped articles, verbless fragments, arrows and made-up abbreviations. "Parser rejects bad date → exit 2, no write" becomes "The parser rejects a bad date, exits with code 2, and writes nothing."

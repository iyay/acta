---
name: routing-chat-01
description: A plain question in Indonesian about git branches should get a chat answer, no files or workflow.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

Bang, mau nanya soal git. Kalau saya sudah bikin branch baru dari main,
terus ada commit baru masuk ke main, bagaimana cara membawa commit itu ke
branch saya tanpa merusak riwayat? Jelaskan saja, tidak usah ubah apa pun
di repo ini.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

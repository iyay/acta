---
name: routing-chat-03
description: A plain question in Indonesian about slow queries should get a chat answer, no files or workflow.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: [Skill]
---

Mau nanya soal Postgres. Query laporan saya lambat sekali begitu datanya
mulai besar, padahal kolom yang dipakai buat saring sudah ada indeksnya.
Apa saja yang biasanya dicek duluan, dan bagaimana cara membaca hasilnya?
Jelaskan saja, tidak usah ubah apa pun di repo ini.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

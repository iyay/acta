---
name: routing-light-02
description: A remark in Indonesian about one wrong word in a config value should take the light path.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

Ada satu nilai yang keliru di config/app.yaml. Kunci timeout tertulis 30,
padahal seharusnya 300, jadi setiap permintaan gagal terlalu cepat.
Perbaikannya satu angka di satu baris itu saja.

The file holds one wrong number and nothing else is broken. Apakah
perubahan ini perlu proses panjang dengan beberapa pendekatan dan
tulisan berseksi, atau rute singkat sudah cukup? Jawab dalam beberapa
baris lalu berhenti. Jangan ubah file apa pun.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

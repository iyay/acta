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

The file holds one wrong number and nothing else in there is broken
at all. Betulkan satu angka di satu baris itu saja.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

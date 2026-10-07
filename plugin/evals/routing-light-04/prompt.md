---
name: routing-light-04
description: A remark in Indonesian about one wrong column name in a query should take the light path.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

Satu nama kolom di berkas queries/orders.sql keliru. Tertulis
customer_nam, padahal tabelnya memakai customer_name, jadi laporan
pesanan selalu gagal. Perbaikannya satu kata di satu baris itu saja.

Only one name is wrong and the rest of the query is fine. Apakah
perubahan ini perlu proses panjang dengan beberapa pendekatan dan
tulisan berseksi, atau rute singkat sudah cukup? Jawab dalam beberapa
baris lalu berhenti. Jangan ubah file apa pun.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

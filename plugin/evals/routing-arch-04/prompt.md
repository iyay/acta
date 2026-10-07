---
name: routing-arch-04
description: A remark in Indonesian about splitting the app into services should start an Architectural brainstorm.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

Saya mau aplikasi ini dipecah menjadi beberapa layanan. Sekarang semuanya
berjalan dalam satu proses dan saling memanggil langsung, dan pemecahan
ini akan mengubah cara setiap bagian berkomunikasi, jadi ini bukan
perbaikan kecil.

Pemecahan ini mengubah setiap panggilan di antara bagian-bagian itu,
jadi tidak ada yang tetap seperti semula. Mulai kerjakan sekarang
dengan cara repo ini menangani perubahan seperti ini, dan berhenti
di hal pertama yang diminta ke saya.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

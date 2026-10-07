---
name: routing-arch-02
description: A remark in Indonesian about moving storage to another database should start an Architectural brainstorm.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

Kami ingin memindah penyimpanan aplikasi ini ke Postgres. Sekarang setiap
modul baca tulis dengan caranya sendiri, dan pemindahan ini akan mengubah
cara semuanya menyimpan data, jadi ini bukan perbaikan kecil.

Semua baca tulis tersebar di setiap modul, jadi pemindahan ini menyentuh
semuanya. Mulai kerjakan sekarang dengan cara repo ini menangani
perubahan seperti ini, dan berhenti di hal pertama yang diminta ke saya.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.

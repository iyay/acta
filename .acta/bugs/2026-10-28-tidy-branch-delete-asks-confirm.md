---
id: BUG-0056
hash: dlm6sip
fixed_in: a683984eccfc6f19b865400f640ea551696b3198
finished: "2026-10-28 09:24:07"
---
# Land selalu nanya sebelum hapus branch

## Symptom
Setiap land nanya "Sisa branch lokal X. Penghapusan paksa akan menghilangkan referensi branch itu. Boleh dihapus?" Padahal land harusnya otomatis bersihin worktree dan branch tanpa nanya.

## Root cause
plugin/hooks/default-rules.md:8 ("Never run a destructive command without a full-sentence warning and a yes", sumber internal/hook/hook.go:57) vs plugin/skills/land/tidy.md:7 (cleanup wajib `git branch -D`). Agent baca `-D` sebagai destructive, jadi nanya. Default commit_history tidy bikin semua land kena ini; mode full pakai `-d` yang safe.

## Repro
Land plan apapun dengan commit_history tidy: worktree kehapus, branch nyangkut, muncul prompt konfirmasi hapus.

## Found in
main, ditemukan saat baca kode setelah user lapor prompt konfirmasi tiap land.

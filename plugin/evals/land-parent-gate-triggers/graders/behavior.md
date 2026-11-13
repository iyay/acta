---
# guards: Apply every slow-gate selection rule from step 1 to those paths
# Scripts append only completed calls; reading or quoting a command cannot pass.
type: regex
pattern: '^(?=[\s\S]*\nfull-A full\n)(?=[\s\S]*\nfull-A slow --area=unit\n)(?=[\s\S]*\nfull-B full\n)(?=[\s\S]*\nfull-B slow --all\n)(?=[\s\S]*\nfull-C full\n)(?=[\s\S]*\nfull-D full\n)(?=[\s\S]*\nfull-D slow --all\n)(?=[\s\S]*\nfull-E full\n)(?=[\s\S]*\nfull-E slow --all\n)(?=[\s\S]*\nfull-F full\n)(?=[\s\S]*\nfull-F slow --all\n)(?=[\s\S]*\ntidy-A full\n)(?=[\s\S]*\ntidy-A slow --area=unit\n)(?=[\s\S]*\ntidy-B full\n)(?=[\s\S]*\ntidy-B slow --all\n)(?=[\s\S]*\ntidy-C full\n)(?=[\s\S]*\ntidy-D full\n)(?=[\s\S]*\ntidy-D slow --all\n)(?=[\s\S]*\ntidy-E full\n)(?=[\s\S]*\ntidy-E slow --all\n)(?=[\s\S]*\ntidy-F full\n)(?=[\s\S]*\ntidy-F slow --all\n)(?![\s\S]*\nfull-A slow (?!--area=unit\n))(?![\s\S]*\nfull-B slow (?!--all\n))(?![\s\S]*\nfull-C slow )(?![\s\S]*\nfull-D slow (?!--all\n))(?![\s\S]*\nfull-E slow (?!--all\n))(?![\s\S]*\nfull-F slow (?!--all\n))(?![\s\S]*\ntidy-A slow (?!--area=unit\n))(?![\s\S]*\ntidy-B slow (?!--all\n))(?![\s\S]*\ntidy-C slow )(?![\s\S]*\ntidy-D slow (?!--all\n))(?![\s\S]*\ntidy-E slow (?!--all\n))(?![\s\S]*\ntidy-F slow (?!--all\n))'
target:
  source: file
  path: calls.log
---

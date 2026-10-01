[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$source = $env:FLOWAK_REHEARSAL_SOURCE_DATABASE_URL
$target = $env:FLOWAK_REHEARSAL_DATABASE_URL
if (-not $source -or -not $target) { throw 'Set FLOWAK_REHEARSAL_SOURCE_DATABASE_URL and FLOWAK_REHEARSAL_DATABASE_URL.' }
if ($source -match '(?i)prod(uction)?' -or $target -match '(?i)prod(uction)?') { throw 'Migration rehearsal refuses production-looking database URLs.' }
if ($source -eq $target) { throw 'Source and target rehearsal databases must differ.' }
foreach ($command in 'pg_dump', 'pg_restore', 'psql') { if (-not (Get-Command $command -ErrorAction SilentlyContinue)) { throw "$command is required for rehearsal." } }

$evidence = Join-Path $root 'tmp\migration-rehearsal'
New-Item -ItemType Directory -Force -Path $evidence | Out-Null
$backup = Join-Path $evidence 'source.dump'
$report = Join-Path $evidence 'report.md'

& pg_dump --format=custom --file $backup $source
$backupHash = (Get-FileHash -Algorithm SHA256 $backup).Hash
$counts = "SELECT table_name, (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name), false, true, ''))[1]::text::bigint) AS rows FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' ORDER BY table_name;"
$sourceCounts = & psql $source -At -F '|' -c $counts
& pg_restore --clean --if-exists --no-owner --dbname $target $backup
Get-ChildItem (Join-Path $root 'backend\db\migrations\*.sql') | Sort-Object Name | ForEach-Object { & psql $target -v ON_ERROR_STOP=1 -f $_.FullName }
$targetCounts = & psql $target -At -F '|' -c $counts
if (($sourceCounts -join "`n") -ne ($targetCounts -join "`n")) { throw 'Row-count mismatch after rehearsal; inspect backup and reconciliation before release.' }
$schemaChecksum = (& psql $target -At -c "SELECT md5(string_agg(table_name || ':' || column_name || ':' || data_type, '|' ORDER BY table_name, ordinal_position)) FROM information_schema.columns WHERE table_schema='public';")
@"
# Migration rehearsal report

- Source/target: non-production operator-provided URLs (redacted)
- Backup SHA-256: `$backupHash
- Schema checksum: `$schemaChecksum
- Row counts: matched
- Migration files: applied in lexical order
- Rollback record: restore `$backup to a new disposable database and repoint the deployment; do not edit old migrations.
"@ | Set-Content -NoNewline $report
Write-Host "Rehearsal passed. Attach $report and the backup manifest to the PR."

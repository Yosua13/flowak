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

function Invoke-CheckedNative {
    param(
        [Parameter(Mandatory)][string]$Command,
        [Parameter(Mandatory)][string[]]$Arguments
    )
    $output = & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command failed with exit code $LASTEXITCODE." }
    return $output
}

$evidence = Join-Path $root 'tmp\migration-rehearsal'
New-Item -ItemType Directory -Force -Path $evidence | Out-Null
$backup = Join-Path $evidence 'source.dump'
$report = Join-Path $evidence 'report.md'

Invoke-CheckedNative -Command 'pg_dump' -Arguments @('--format=custom', '--file', $backup, $source) | Out-Null
$backupHash = (Get-FileHash -Algorithm SHA256 $backup).Hash
$counts = "SELECT table_name, ((xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name), false, true, '')))[1]::text)::bigint AS rows FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' ORDER BY table_name;"
$sourceCounts = Invoke-CheckedNative -Command 'psql' -Arguments @($source, '-At', '-F', '|', '-v', 'ON_ERROR_STOP=1', '-c', $counts)
Invoke-CheckedNative -Command 'pg_restore' -Arguments @('--clean', '--if-exists', '--no-owner', '--exit-on-error', '--dbname', $target, $backup) | Out-Null
$migrations = Get-ChildItem (Join-Path $root 'backend\db\migrations\*.sql') | Sort-Object Name
$migrations | ForEach-Object { Invoke-CheckedNative -Command 'psql' -Arguments @($target, '-v', 'ON_ERROR_STOP=1', '-f', $_.FullName) | Out-Null }
$targetCounts = Invoke-CheckedNative -Command 'psql' -Arguments @($target, '-At', '-F', '|', '-v', 'ON_ERROR_STOP=1', '-c', $counts)
if (($sourceCounts -join "`n") -ne ($targetCounts -join "`n")) { throw 'Row-count mismatch after rehearsal; inspect backup and reconciliation before release.' }
$schemaChecksum = Invoke-CheckedNative -Command 'psql' -Arguments @($target, '-At', '-v', 'ON_ERROR_STOP=1', '-c', "SELECT md5(string_agg(table_name || ':' || column_name || ':' || data_type, '|' ORDER BY table_name, ordinal_position)) FROM information_schema.columns WHERE table_schema='public';")
$rowCountEvidence = ($targetCounts | ForEach-Object { "  - $_" }) -join "`n"
@"
# Migration rehearsal report

- Source/target: non-production operator-provided URLs (redacted)
- Backup SHA-256: $backupHash
- Schema checksum: $schemaChecksum
- Row counts: matched
- Migration files: $($migrations.Count) applied in lexical order
- Restore verification: backup restored into a distinct disposable target before migrations were replayed
- Rollback record: restore the verified backup into a new disposable database and repoint the deployment; do not edit old migrations.

## Row-count snapshot

$rowCountEvidence
"@ | Set-Content -NoNewline $report
Write-Host "Rehearsal passed. Attach $report and the backup manifest to the PR."

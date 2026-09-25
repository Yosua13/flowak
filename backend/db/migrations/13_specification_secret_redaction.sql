-- Forward-only repair for legacy specification snapshots. Keep non-secret data,
-- replace credential-bearing fields with references/placeholders before reads.
CREATE OR REPLACE FUNCTION flowak_scrub_spec_json(value JSONB) RETURNS JSONB
LANGUAGE plpgsql AS $$
DECLARE key TEXT; child JSONB; result JSONB; raw TEXT; parsed JSONB;
BEGIN
  IF value IS NULL THEN RETURN NULL; END IF;
  IF jsonb_typeof(value) = 'object' THEN
    result := '{}'::jsonb;
    FOR key, child IN SELECT * FROM jsonb_each(value) LOOP
      IF key ~* '(authorization|cookie|password|secret|token|api[_-]?key)' OR key = 'curl' THEN
        result := result || jsonb_build_object(key, '{{REDACTED}}');
      ELSIF key = 'auth' THEN
        raw := child #>> '{}';
        IF raw ~ '^\{\{[A-Z][A-Z0-9_]*\}\}$' OR raw IN ('none', 'inherit') THEN
          result := result || jsonb_build_object(key, child);
        ELSE
          result := result || jsonb_build_object(key, '{{API_TOKEN}}');
        END IF;
      ELSE
        result := result || jsonb_build_object(key, flowak_scrub_spec_json(child));
      END IF;
    END LOOP;
    RETURN result;
  ELSIF jsonb_typeof(value) = 'array' THEN
    SELECT COALESCE(jsonb_agg(flowak_scrub_spec_json(item) ORDER BY position), '[]'::jsonb)
      INTO result FROM jsonb_array_elements(value) WITH ORDINALITY AS entries(item, position);
    RETURN result;
  ELSIF jsonb_typeof(value) = 'string' THEN
    raw := value #>> '{}';
    IF left(ltrim(raw), 1) IN ('{', '[') THEN
      BEGIN
        parsed := raw::jsonb;
        RETURN to_jsonb(flowak_scrub_spec_json(parsed)::text);
      EXCEPTION WHEN invalid_text_representation THEN NULL;
      END;
    END IF;
    RETURN to_jsonb(regexp_replace(raw, '(?i)Bearer[[:space:]]+[^[:space:]"'']+', 'Bearer {{REDACTED}}', 'g'));
  END IF;
  RETURN value;
END $$;

UPDATE modules SET nodes = flowak_scrub_spec_json(nodes::jsonb)::text WHERE nodes IS NOT NULL;
UPDATE workflow_nodes SET metadata = flowak_scrub_spec_json(metadata)
  || jsonb_build_object('legacy_notes', COALESCE(metadata->'legacy_notes', '{}'::jsonb)
    || CASE WHEN metadata #>> '{raw,doc,sla}' IS NOT NULL AND metadata #>> '{raw,doc,sla}' !~* '^\s*\d+(?:[.,]\d+)?\s*(menit|minute|minutes|jam|hour|hours|hari|day|days|minggu|week|weeks)\s*$'
      THEN jsonb_build_object('sla', flowak_scrub_spec_json(metadata #> '{raw,doc,sla}')) ELSE '{}'::jsonb END
    || CASE WHEN metadata #>> '{raw,roles,backend,auth}' IS NOT NULL AND metadata #>> '{raw,roles,backend,auth}' !~ '^\{\{[A-Z][A-Z0-9_]*\}\}$'
      THEN jsonb_build_object('backend.auth', 'Legacy authorization removed; choose a variable reference.') ELSE '{}'::jsonb END);
UPDATE workflow_edges SET metadata = flowak_scrub_spec_json(metadata);
UPDATE node_role_tasks SET metadata = flowak_scrub_spec_json(metadata);
UPDATE node_uiux_specs SET metadata = flowak_scrub_spec_json(metadata);
UPDATE node_frontend_specs SET metadata = flowak_scrub_spec_json(metadata);
UPDATE node_api_contracts SET
  auth_policy = CASE WHEN auth_policy IS NULL OR auth_policy = '' OR auth_policy ~ '^\{\{[A-Z][A-Z0-9_]*\}\}$' OR auth_policy IN ('none', 'inherit') THEN auth_policy ELSE '{{API_TOKEN}}' END,
  curl_example = CASE WHEN curl_example IS NULL THEN NULL ELSE '{{REDACTED}}' END,
  request_example = flowak_scrub_spec_json(request_example),
  response_example = flowak_scrub_spec_json(response_example),
  error_codes = flowak_scrub_spec_json(error_codes),
  metadata = flowak_scrub_spec_json(metadata);
UPDATE module_versions SET graph_snapshot = flowak_scrub_spec_json(graph_snapshot), diff_summary = flowak_scrub_spec_json(diff_summary);
UPDATE activity_logs SET before_data = flowak_scrub_spec_json(before_data), after_data = flowak_scrub_spec_json(after_data);
UPDATE event_outbox SET payload = flowak_scrub_spec_json(payload);
UPDATE notifications SET payload = flowak_scrub_spec_json(payload);

DROP FUNCTION flowak_scrub_spec_json(JSONB);

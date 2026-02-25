DECLARE
  v_offset INT;
BEGIN
  IF p_program_id IS NULL THEN
    RAISE EXCEPTION 'p_program_id is required';
  END IF;

  v_offset := COALESCE(p_page, 0) * COALESCE(p_page_size, 10);

  RETURN QUERY
  WITH org_filtered_users AS (
    SELECT DISTINCT ua.user_id
    FROM user_attributes ua
    WHERE ua.attribute_category = 'employee'
      AND ua.attribute_key = 'organization'
      AND (p_organization_id IS NULL OR ua.attribute_value = p_organization_id::TEXT)
  ),
  dept_filtered_users AS (
    SELECT DISTINCT ua.user_id
    FROM user_attributes ua
    WHERE ua.attribute_category = 'employee'
      AND ua.attribute_key = 'department'
      AND (p_department_id IS NULL OR ua.attribute_value = p_department_id::TEXT)
  ),
  section_filtered_users AS (
    SELECT DISTINCT ua.user_id
    FROM user_attributes ua
    WHERE ua.attribute_category = 'employee'
      AND ua.attribute_key = 'section'
      AND (p_section_id IS NULL OR ua.attribute_value = p_section_id::TEXT)
  ),
  job_filtered_users AS (
    SELECT DISTINCT ua.user_id
    FROM user_attributes ua
    WHERE ua.attribute_category = 'employee'
      AND ua.attribute_key = 'job_position'
      AND (
        p_job_position_ids IS NULL
        OR array_length(p_job_position_ids, 1) IS NULL
        OR ua.attribute_value = ANY(ARRAY(SELECT uuid::text FROM unnest(p_job_position_ids) AS uuid))
      )
  ),
  grade_filtered_users AS (
    SELECT DISTINCT ua.user_id
    FROM user_attributes ua
    WHERE ua.attribute_category = 'employee'
      AND ua.attribute_key = 'job_level'
      AND (
        p_grade_ids IS NULL
        OR array_length(p_grade_ids, 1) IS NULL
        OR ua.attribute_value = ANY(ARRAY(SELECT uuid::text FROM unnest(p_grade_ids) AS uuid))
      )
  ),
  filtered_user_ids AS (
    SELECT ofu.user_id
    FROM org_filtered_users ofu
    INNER JOIN dept_filtered_users dfu ON dfu.user_id = ofu.user_id
    INNER JOIN section_filtered_users sfu ON sfu.user_id = ofu.user_id
    INNER JOIN job_filtered_users jfu ON jfu.user_id = ofu.user_id
    INNER JOIN grade_filtered_users gfu ON gfu.user_id = ofu.user_id
    WHERE (
      p_excluded_user_ids IS NULL
      OR array_length(p_excluded_user_ids, 1) IS NULL
      OR ofu.user_id != ALL(p_excluded_user_ids)
    )
  ),
  user_profiles_with_attrs AS (
    SELECT 
      up.id,
      up.user_id,
      up.name,
      up.email,
      up.nrp,
      MAX(CASE WHEN ua.attribute_key = 'organization' THEN ua.attribute_value END) AS org_id,
      MAX(CASE WHEN ua.attribute_key = 'department' THEN ua.attribute_value END) AS dept_id,
      MAX(CASE WHEN ua.attribute_key = 'job_position' THEN ua.attribute_value END) AS job_id,
      MAX(CASE WHEN ua.attribute_key = 'job_level' THEN ua.attribute_value END) AS grade_id
    FROM user_profile up
    INNER JOIN filtered_user_ids fui ON up.user_id = fui.user_id
    LEFT JOIN user_attributes ua ON up.user_id = ua.user_id 
      AND ua.attribute_category = 'employee'
      AND ua.attribute_key IN ('organization', 'department', 'job_position', 'job_level')
    WHERE (
      p_search IS NULL
      OR char_length(p_search) < 3
      OR up.name ILIKE CONCAT('%', p_search, '%')
      OR up.nrp ILIKE CONCAT('%', p_search, '%')
    )
    GROUP BY up.id, up.user_id, up.name, up.email, up.nrp
  ),
  enriched_profiles AS (
    SELECT 
      upa.id,
      upa.user_id,
      upa.name,
      upa.email,
      upa.nrp,
      COALESCE(mo.name, upa.org_id) AS organization,
      COALESCE(mou.name, upa.dept_id) AS department,
      COALESCE(mj.name, upa.job_id) AS job_position,
      CASE 
        WHEN mg.name IS NOT NULL THEN CONCAT(mg.name, ' (', COALESCE(mg.label, ''), ')')
        ELSE upa.grade_id
      END AS grade
    FROM user_profiles_with_attrs upa
    LEFT JOIN master_organizations mo 
      ON mo.id = (
        CASE WHEN upa.org_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
             THEN upa.org_id::UUID ELSE NULL END)
    LEFT JOIN master_organization_units mou 
      ON mou.id = (
        CASE WHEN upa.dept_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
             THEN upa.dept_id::UUID ELSE NULL END)
    LEFT JOIN master_job_positions mjp 
      ON mjp.id = (
        CASE WHEN upa.job_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
             THEN upa.job_id::UUID ELSE NULL END)
    LEFT JOIN master_jobs mj ON mjp.job_id = mj.id
    LEFT JOIN master_job_position_grades mjpg 
      ON mjpg.id = (
        CASE WHEN upa.grade_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
             THEN upa.grade_id::UUID ELSE NULL END)
    LEFT JOIN master_grades mg ON mjpg.grade_id = mg.id
  ),
  counted_profiles AS (
    SELECT ep.*, COUNT(*) OVER() AS total_count
    FROM enriched_profiles ep
  )
  SELECT 
    cp.id,
    cp.user_id,
    cp.name::text,
    cp.email::text,
    cp.nrp::text,
    cp.organization::text,
    cp.department::text,
    cp.job_position::text,
    cp.grade::text,
    cp.total_count
  FROM counted_profiles cp
  ORDER BY cp.name ASC NULLS LAST
  OFFSET v_offset
  LIMIT p_page_size;
END;

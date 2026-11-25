DELETE FROM assignment_submissions older
USING assignment_submissions newer
WHERE older.assignment_id = newer.assignment_id
  AND older.student_id = newer.student_id
  AND older.created_at > newer.created_at;

CREATE UNIQUE INDEX IF NOT EXISTS uq_assignment_submissions_assignment_student
ON assignment_submissions (assignment_id, student_id);
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import {
  Video, BookOpen, ClipboardList, Brain, CalendarCheck,
  Users, GraduationCap, BookMarked, Layers, LayoutGrid,
  Plus, Edit2, Trash2, Eye, Download, Upload, X, Check,
  ChevronDown, ChevronUp, Clock, Calendar, FileText,
  ToggleLeft, ToggleRight, CheckCircle2, XCircle, AlertCircle,
  Paperclip, ExternalLink, Award, BarChart2, Activity,
  Mail, Phone, MapPin, Copy, CheckCheck, User, Shield, Hash
} from 'lucide-vue-next';

import { useApi } from "~/composables/useApi";
import { useAuthStore } from "~/stores/auth";

definePageMeta({ middleware: "auth" });

interface Item {
  id?: string;
  title?: string;
  name?: string;
  username?: string;
  email?: string;
  description?: string;
  deadline?: string;
  created_at?: string;
  status?: string;
  role?: { name?: string };
  attendance_date?: string;
  start_time?: string;
  end_time?: string;
  [key: string]: any;
}
type Tab = { key: string; label: string; endpoint: string; create?: boolean };

const auth = useAuthStore();
const { request } = useApi();
const route = useRoute();
const tabsByRole: Record<string, Tab[]> = {
  admin: [
    { key: "meetings", label: "Monitoring meeting", endpoint: "/admin/meetings" },
    { key: "teachers", label: "Guru", endpoint: "/admin/users", create: true },
    { key: "students", label: "Siswa", endpoint: "/admin/users", create: true },
    {
      key: "classes",
      label: "Kelas",
      endpoint: "/admin/classes",
      create: true,
    },
    {
      key: "majors",
      label: "Jurusan",
      endpoint: "/admin/majors",
      create: true,
    },
    {
      key: "subjects",
      label: "Mata pelajaran",
      endpoint: "/admin/subjects",
      create: true,
    },
  ],
  guru: [
    { key: "meetings", label: "Meeting", endpoint: "/meetings" },
    {
      key: "materials",
      label: "Materi",
      endpoint: "/teacher/materials",
      create: true,
    },
    {
      key: "assignments",
      label: "Tugas",
      endpoint: "/teacher/assignments",
      create: true,
    },
    {
      key: "quizzes",
      label: "Kuis",
      endpoint: "/teacher/quizzes",
      create: true,
    },
    {
      key: "attendance",
      label: "Presensi",
      endpoint: "/teacher/attendance-sessions",
      create: true,
    },
  ],
  siswa: [
    { key: "meetings", label: "Meeting", endpoint: "/meetings" },
    { key: "materials", label: "Materi", endpoint: "/student/materials" },
    { key: "assignments", label: "Tugas", endpoint: "/student/assignments" },
    { key: "quizzes", label: "Kuis", endpoint: "/student/quizzes" },
    {
      key: "attendance",
      label: "Presensi",
      endpoint: "/student/attendance-sessions",
    },
  ],
};

/** Reads the first present value from a list of dotted paths.
 *  Consolidates the snake_case / PascalCase duplication coming from the backend
 *  (e.g. `username` vs `Username`, `Student.FullName` vs `student.full_name`). */
const pick = (item: any, ...paths: string[]): any => {
  for (const path of paths) {
    const value = path
      .split(".")
      .reduce((acc, key) => (acc == null ? undefined : acc[key]), item);
    if (value !== undefined && value !== null && value !== "") return value;
  }
  return undefined;
};

const tabs = computed(() => tabsByRole[auth.role] || []);
const activeTab = ref("");
const activeConfig = computed(
  () => tabs.value.find((tab) => tab.key === activeTab.value) || tabs.value[0],
);

const TAB_META: Record<string, { icon: any; color: string; bg: string }> = {
  meetings:    { icon: Video,         color: '#7c3aed', bg: 'rgba(124,58,237,0.12)' },
  materials:   { icon: BookOpen,      color: '#0ea5e9', bg: 'rgba(14,165,233,0.12)' },
  assignments: { icon: ClipboardList, color: '#f59e0b', bg: 'rgba(245,158,11,0.12)' },
  quizzes:     { icon: Brain,         color: '#ed7864', bg: 'rgba(237,120,100,0.12)' },
  attendance:  { icon: CalendarCheck, color: '#5eb69d', bg: 'rgba(94,182,157,0.12)' },
  teachers:    { icon: GraduationCap, color: '#5eb69d', bg: 'rgba(94,182,157,0.12)' },
  students:    { icon: Users,         color: '#0ea5e9', bg: 'rgba(14,165,233,0.12)' },
  classes:     { icon: Layers,        color: '#7c3aed', bg: 'rgba(124,58,237,0.12)' },
  majors:      { icon: BookMarked,    color: '#f59e0b', bg: 'rgba(245,158,11,0.12)' },
  subjects:    { icon: LayoutGrid,    color: '#ed7864', bg: 'rgba(237,120,100,0.12)' },
};
const tabIcon = (key: string) => TAB_META[key]?.icon || Activity;
const tabColor = (key: string) => TAB_META[key]?.color || '#5eb69d';
const tabBg = (key: string) => TAB_META[key]?.bg || 'rgba(94,182,157,0.12)';

const items = ref<Item[]>([]);
const loading = ref(false);
const saving = ref(false);
const errorMessage = ref("");
const errorDetails = ref<{
  reason?: string;
  dependency?: number;
  next_steps?: string[];
} | null>(null);
const showErrorDetails = ref(false);
const successMessage = ref("");
const submissionRows = ref<any[]>([]);
const submissionAssignmentTitle = ref("");
const submissionAssignment = ref<any | null>(null);
const selectedSubmission = ref<any | null>(null);
const gradeForm = ref({ score: "", feedback: "" });
const grading = ref(false);
const showForm = ref(false);
const creationFile = ref<File | null>(null);
// Per-item selected file, keyed by item id, so choosing a file in one row
// can never be submitted/uploaded against a different row.
const selectedFiles = reactive<Record<string, File | null>>({});
const editingId = ref("");
const classSubjectOptions = ref<any[]>([]);
const academicYearOptions = ref<any[]>([]);
const majorOptions = ref<any[]>([]);
const teacherOptions = ref<any[]>([]);
const selectedDetail = ref<any | null>(null);
const studentSubmission = ref<any | null>(null);
const detailQuery = ref("");
const sortBy = ref("name");
const sortDirection = ref("asc");
const classFilter = ref("");
const studentAttendanceStats = ref<{
  total_sessions: number;
  attended_sessions: number;
  excused_sessions: number;
  absent_sessions: number;
  attendance_percentage: number;
} | null>(null);

const attendanceStatusLabel = (status?: string) => {
  switch (status) {
    case "present":
      return "Hadir";
    case "late":
      return "Terlambat";
    case "excused":
      return "Izin";
    case "absent":
      return "Alpa";
    default:
      return "Sudah Hadir";
  }
};

const toggleAttendanceSession = async (item: Item) => {
  const id = itemId(item);
  if (!id) return;
  const nextState = item.status === "open" ? "close" : "open";
  try {
    await request(`/teacher/attendance-sessions/${encodeURIComponent(id)}/${nextState}`, {
      method: "POST",
    });
    item.status = nextState === "open" ? "open" : "closed";
    successMessage.value =
      nextState === "open"
        ? "Sesi presensi berhasil dibuka."
        : "Sesi presensi ditutup.";
    setTimeout(() => {
      successMessage.value = "";
    }, 3000);
  } catch (error: any) {
    setApiError(error, "Gagal mengubah status sesi presensi.");
  }
};

// Which sort fields make sense depends on what the current tab actually has.
// "Metadata" as a catch-all sort was mostly meaningless for tabs like classes/majors/subjects.
const sortOptions = computed(() => {
  const options = [{ value: "name", label: "Nama" }];
  if (activeTab.value === "teachers")
    options.push({ value: "meta", label: "Email" });
  if (activeTab.value === "students")
    options.push(
      { value: "class", label: "Kelas" },
      { value: "meta", label: "Email" },
    );
  if (activeTab.value === "classes")
    options.push({ value: "grade", label: "Tingkat" });
  if (["majors", "subjects"].includes(activeTab.value))
    options.push({ value: "code", label: "Kode" });
  return options;
});
// Guards against out-of-order responses when openDetail is called repeatedly in quick succession.
let detailRequestToken = 0;

const form = ref<Record<string, any>>({});
const title = computed(() =>
  auth.role === "admin"
    ? "Administrasi sekolah"
    : auth.role === "guru"
      ? "Ruang kerja guru"
      : "Ruang belajar",
);
const isTeacher = computed(() => auth.role === "guru");
const isStudent = computed(() => auth.role === "siswa");
const isTeacherForm = computed(() => activeTab.value === "teachers");
const isStudentForm = computed(() => activeTab.value === "students");
const optionId = (item: any) => pick(item, "id", "ID");
const optionName = (item: any) =>
  pick(
    item,
    "name",
    "Name",
    "title",
    "Title",
    "full_name",
    "FullName",
    "username",
    "Username",
  ) || "Tanpa nama";
const itemClassNames = (item: any) =>
  (pick(item, "Student.Classes", "student.classes") || []).map(
    (classItem: any) => optionName(classItem),
  );
const classFilterOptions = computed(() =>
  Array.from(new Set(items.value.flatMap(itemClassNames))).sort(),
);

const sortValue = (item: any): string | number => {
  switch (sortBy.value) {
    case "class":
      return itemClassNames(item).join(", ");
    case "meta":
      return itemMeta(item);
    case "code":
      return pick(item, "code", "Code") || "";
    case "grade":
      return Number(pick(item, "grade", "Grade") ?? 0);
    default:
      return itemTitle(item);
  }
};

const displayedItems = computed(() => {
  const query = detailQuery.value.trim().toLowerCase();
  const filtered = items.value.filter((item) => {
    const matchesText =
      !query ||
      `${itemTitle(item)} ${itemMeta(item)} ${itemDescription(item)} ${itemClassNames(item).join(" ")}`
        .toLowerCase()
        .includes(query);
    const matchesClass =
      !classFilter.value || itemClassNames(item).includes(classFilter.value);
    return matchesText && matchesClass;
  });
  const direction = sortDirection.value === "asc" ? 1 : -1;
  return [...filtered].sort((left, right) => {
    if (sortBy.value === "grade")
      return (
        ((sortValue(left) as number) - (sortValue(right) as number)) * direction
      );
    return (
      String(sortValue(left))
        .toLowerCase()
        .localeCompare(String(sortValue(right)).toLowerCase()) * direction
    );
  });
});

const classSubjectName = (item: any) => {
  const classItem = pick(item, "class", "Class");
  const subject = pick(item, "subject", "Subject");
  return `${optionName(classItem) || "Kelas"} · ${optionName(subject) || "Mata pelajaran"}`;
};

const addClassSubject = (item: any) => {
  const id = optionId(item) || item.class_subject_id || item.ClassSubjectID;
  if (
    !id ||
    classSubjectOptions.value.some((option) => optionId(option) === id)
  )
    return;
  classSubjectOptions.value.push({
    ...item,
    id,
    label: classSubjectName(item),
  });
};

const loadReferenceData = async () => {
  try {
    if (auth.role === "admin") {
      const [
        classResponse,
        subjectResponse,
        yearResponse,
        majorResponse,
        userResponse,
      ] = await Promise.all([
        request<{ data?: any[] }>("/admin/classes"),
        request<{ data?: any[] }>("/admin/subjects"),
        request<{ data?: any[] }>("/admin/academic-years"),
        request<{ data?: any[] }>("/admin/majors"),
        request<{ data?: any[] }>("/admin/users"),
      ]);
      academicYearOptions.value = Array.from(
        new Map<string, any>(
          (classResponse.data || [])
            .map((item: any): [string, any] => {
              const year = pick(item, "academic_year", "AcademicYear");
              return [String(optionId(year)), year] as [string, any];
            })
            .filter(([id]) => id),
        ).values(),
      );
      if (yearResponse.data?.length)
        academicYearOptions.value = yearResponse.data;
      majorOptions.value = majorResponse.data || [];
      teacherOptions.value = (userResponse.data || []).filter(
        (item: any) =>
          (
            pick(item, "role.name", "role.Name", "Role.Name", "Role.name") ||
            ""
          ).toLowerCase() === "guru",
      );

      // Fetch every class's subject detail in parallel instead of one request at a time.
      const classDetails = await Promise.all(
        (classResponse.data || []).map((item: any) => {
          const classId = optionId(item);
          if (!classId) return Promise.resolve(null);
          return request<{ data?: any[] }>(
            `/admin/classes/${classId}/subjects/detail`,
          ).catch(() => null);
        }),
      );
      for (const detail of classDetails)
        for (const classSubject of detail?.data || [])
          addClassSubject(classSubject);
      for (const subject of subjectResponse.data || [])
        addClassSubject({ Subject: subject });
    } else {
      const endpoints =
        auth.role === "guru"
          ? ["/teacher/class-subjects"]
          : [
              "/student/materials",
              "/student/assignments",
              "/student/quizzes",
              "/student/attendance-sessions",
            ];
      const responses = await Promise.all(
        endpoints.map((endpoint) =>
          request<{ data?: any[] }>(endpoint).catch(() => ({ data: [] })),
        ),
      );
      for (const response of responses)
        for (const item of response.data || [])
          addClassSubject(pick(item, "class_subject", "ClassSubject") || item);
    }
  } catch {
    /* Reference dropdowns remain empty when the role has no source data. */
  }
};

const resetForm = (item?: Item) => {
  editingId.value = item ? itemId(item) || "" : "";
  creationFile.value = null;
  form.value = ["teachers", "students"].includes(activeTab.value)
    ? {
        username: "",
        password: "",
        email: "",
        role: activeTab.value === "teachers" ? "guru" : "siswa",
        full_name: "",
        nip: "",
        phone: "",
        nis: "",
        nisn: "",
        gender: "",
        address: "",
      }
    : activeTab.value === "classes"
      ? {
          name: "",
          grade: 10,
          major_id: "",
          academic_year_id: academicYearOptions.value[0]
            ? optionId(academicYearOptions.value[0])
            : "",
        }
      : activeTab.value === "majors"
        ? { code: "", name: "", description: "" }
        : activeTab.value === "subjects"
          ? { code: "", name: "", description: "", teacher_ids: [] }
          : activeTab.value === "materials"
            ? {
                class_subject_id: "",
                title: "",
                description: "",
                external_url: "",
                video_url: "",
                status: "draft",
              }
            : activeTab.value === "assignments"
              ? {
                  class_subject_id: "",
                  title: "",
                  description: "",
                  deadline: "",
                  max_score: 100,
                }
              : activeTab.value === "quizzes"
                ? {
                    class_subject_id: "",
                    title: "",
                    description: "",
                    duration: 30,
                    max_attempts: 1,
                  }
                : {
                    class_subject_id: "",
                    title: "",
                    description: "",
                    attendance_date: new Date().toISOString().slice(0, 10),
                    start_time: "07:00",
                    end_time: "08:00",
                  };
};

const loadItems = async () => {
  if (!activeConfig.value) return;
  loading.value = true;
  errorMessage.value = "";
  errorDetails.value = null;
  showErrorDetails.value = false;
  try {
    const response = await request<{ data?: Item[] }>(
      activeConfig.value.endpoint,
    );
    const role =
      activeTab.value === "teachers"
        ? "guru"
        : activeTab.value === "students"
          ? "siswa"
          : "";
    items.value = (Array.isArray(response.data) ? response.data : []).filter(
      (item: any) =>
        !role ||
        (
          pick(item, "role.name", "role.Name", "Role.Name", "Role.name") ||
          ""
        )
          .toLowerCase() === role,
    );
    const requestedStudentId = String(route.query.student_id || "");
    if (activeTab.value === "students" && requestedStudentId) {
      const requestedStudent = items.value.find(
        (item) =>
          optionId(item) === requestedStudentId ||
          optionId(pick(item, "Student", "student")) === requestedStudentId,
      );
      if (requestedStudent) await openDetail(requestedStudent);
    }
    if (activeTab.value === "attendance" && isStudent.value) {
      request<{ data?: any }>("/student/attendance-statistics")
        .then((res) => {
          studentAttendanceStats.value = res.data || null;
        })
        .catch(() => {
          studentAttendanceStats.value = null;
        });
    }
  } catch (error: any) {
    items.value = [];
    errorMessage.value = error?.data?.message || "Data belum dapat dimuat.";
  } finally {
    loading.value = false;
  }
};

const setApiError = (error: any, fallback: string) => {
  errorMessage.value = error?.data?.message || fallback;
  errorDetails.value = error?.data?.details || null;
  showErrorDetails.value = false;
};

const openCreate = () => {
  resetForm();
  successMessage.value = "";
  errorMessage.value = "";
  showForm.value = true;
};
const openEdit = (item: Item) => {
  resetForm(item);
  form.value = {
    ...form.value,
    username: pick(item, "username", "Username") || "",
    email: pick(item, "email", "Email") || "",
    full_name: pick(item, "Student.FullName", "Teacher.FullName") || "",
    nip: pick(item, "Teacher.NIP") || "",
    phone: pick(item, "Student.Phone", "Teacher.Phone") || "",
    nis: pick(item, "Student.NIS") || "",
    nisn: pick(item, "Student.NISN") || "",
    gender: pick(item, "Student.Gender") || "",
    address: pick(item, "Student.Address") || "",
    grade: pick(item, "grade", "Grade") || 10,
    major_id: pick(item, "major_id", "MajorID", "Major.ID") || "",
    academic_year_id: pick(item, "academic_year_id", "AcademicYearID") || "",
    code: pick(item, "code", "Code") || "",
    teacher_ids: (pick(item, "Teachers", "teachers") || [])
      .map((teacher: any) => pick(teacher, "id", "ID"))
      .filter(Boolean),
    class_subject_id:
      pick(
        item,
        "class_subject_id",
        "ClassSubjectID",
        "class_subject.id",
        "ClassSubject.ID",
      ) || "",
    title: pick(item, "title", "Title") || "",
    description: pick(item, "description", "Description") || "",
    deadline: item.deadline
      ? new Date(item.deadline).toISOString().slice(0, 16)
      : "",
    max_score: pick(item, "max_score", "MaxScore") || 100,
    duration: pick(item, "duration", "Duration") || 30,
    max_attempts: pick(item, "max_attempts", "MaxAttempts") || 1,
    status: pick(item, "status", "Status") || "draft",
    external_url: pick(item, "external_url", "ExternalURL") || "",
    video_url: pick(item, "video_url", "VideoURL") || "",
  };
  showForm.value = true;
};

const openDetail = async (item: Item) => {
  const resourceId = activeTab.value === "assignments" ? assignmentId(item) : itemId(item);
  if (!activeConfig.value || !resourceId) {
    errorMessage.value = "ID tugas tidak ditemukan. Muat ulang daftar tugas lalu coba lagi.";
    return;
  }
  const requestToken = ++detailRequestToken;
  studentSubmission.value = null;
  errorMessage.value = "";
  try {
    const detailPath =
      activeTab.value === "classes"
        ? `${activeConfig.value.endpoint}/detail/${resourceId}`
        : `${activeConfig.value.endpoint}/${resourceId}`;
    const response = await request<{ data?: any }>(detailPath);
    if (requestToken !== detailRequestToken) return; // a newer request has since started; discard this one
    selectedDetail.value = response.data || item;
    if (
      isTeacher.value &&
      ["materials", "assignments"].includes(activeTab.value) &&
      !resourceFiles(selectedDetail.value).length
    ) {
      const resource =
        activeTab.value === "materials" ? "materials" : "assignments";
      const filesResponse = await request<{ data?: any[] }>(
        `/teacher/${resource}/${itemId(item)}/files`,
      ).catch(() => ({ data: [] }));
      selectedDetail.value = {
        ...selectedDetail.value,
        Files: filesResponse.data || [],
      };
    }
    if (isStudent.value && activeTab.value === "assignments") {
      studentSubmission.value = await request<{ data?: any }>(
        `/student/assignments/${encodeURIComponent(assignmentId(item))}/submission`,
      )
        .then((result) => result.data || null)
        .catch(() => null);
    }
  } catch (error: any) {
    if (requestToken !== detailRequestToken) return;
    selectedDetail.value = item;
    setApiError(error, "Detail tugas belum dapat dibuka. Periksa koneksi atau muat ulang halaman.");
  }
};
const closeForm = () => {
  showForm.value = false;
};
const closeDetail = () => {
  selectedDetail.value = null;
  studentSubmission.value = null;
};

const submitForm = async () => {
  if (!activeConfig.value) return;
  errorMessage.value = "";
  successMessage.value = "";
  if (
    activeTab.value === "attendance" &&
    form.value.start_time &&
    form.value.end_time &&
    form.value.end_time <= form.value.start_time
  ) {
    errorMessage.value = "Jam selesai harus setelah jam mulai.";
    return;
  }
  saving.value = true;
  try {
    const endpoint = editingId.value
      ? activeTab.value === "classes"
        ? `${activeConfig.value.endpoint}/update/${editingId.value}`
        : `${activeConfig.value.endpoint}/${editingId.value}`
      : activeConfig.value.endpoint;
    const method = editingId.value ? "PUT" : "POST";
    const payload = { ...form.value };
    if (activeTab.value === "teachers") payload.role = "guru";
    else if (activeTab.value === "students") payload.role = "siswa";
    else if (activeTab.value === "attendance") {
      payload.start_time = `${form.value.attendance_date} ${form.value.start_time}:00`;
      payload.end_time = `${form.value.attendance_date} ${form.value.end_time}:00`;
    }
    const response = await request<{ data?: any }>(endpoint, {
      method,
      body: payload,
    });
    if (
      !editingId.value &&
      creationFile.value &&
      ["materials", "assignments"].includes(activeTab.value)
    ) {
      const resourceId = itemId(response.data || {});
      if (!resourceId)
        throw new Error("ID resource tidak ditemukan setelah dibuat");
      const fileBody = new FormData();
      fileBody.append("file", creationFile.value);
      const resource =
        activeTab.value === "materials" ? "materials" : "assignments";
      await request(`/teacher/${resource}/${resourceId}/files`, {
        method: "POST",
        body: fileBody,
      });
    }
    successMessage.value = `${activeConfig.value.label} berhasil ${editingId.value ? "diperbarui" : "dibuat"}.`;
    closeForm();
    await loadItems();
  } catch (error: any) {
    setApiError(error, "Data tidak dapat disimpan.");
  } finally {
    saving.value = false;
  }
};

type RequestMethod = "DELETE" | "POST" | "PATCH" | "PUT";
const action = async (
  method: RequestMethod,
  path: string,
  message: string,
) => {
  try {
    await request(path, { method: method as "DELETE" | "POST" | "PATCH" | "PUT" });
    successMessage.value = message;
    await loadItems();
  } catch (error: any) {
    setApiError(error, "Aksi tidak dapat dilakukan.");
  }
};

const itemId = (item: Item) => pick(item, "id", "ID") || "";
const assignmentId = (item: any) =>
  pick(item, "id", "ID", "assignment_id", "AssignmentID") ||
  pick(
    item,
    "Assignment.id",
    "Assignment.ID",
    "assignment.id",
    "assignment.ID",
  ) ||
  "";

// Blocks a new assignment from being created with a deadline already in the past.
// Doesn't apply while editing, since an existing (already-passed) deadline must stay editable.
const minDeadline = computed(() => {
  if (editingId.value) return undefined;
  const now = new Date();
  now.setSeconds(0, 0);
  now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
  return now.toISOString().slice(0, 16);
});
const publish = (item: Item) =>
  itemId(item) &&
  action(
    activeTab.value === "quizzes" ? "POST" : "PATCH",
    `${activeConfig.value?.endpoint}/${itemId(item)}/publish`,
    "Item berhasil dipublikasikan.",
  );
const closeQuiz = (item: Item) =>
  itemId(item) &&
  action(
    "POST",
    `${activeConfig.value?.endpoint}/${itemId(item)}/close`,
    "Kuis berhasil ditutup.",
  );
const remove = (item: Item) => {
  if (!itemId(item) || !confirm("Hapus item ini?")) return;
  const path =
    activeTab.value === "classes" || activeTab.value === "subjects"
      ? `${activeConfig.value?.endpoint}/delete/${itemId(item)}`
      : `${activeConfig.value?.endpoint}/${itemId(item)}`;
  action("DELETE", path, "Item berhasil dihapus.");
};

const chooseFile = (item: Item, event: Event) => {
  const id = itemId(item);
  if (!id) return;
  selectedFiles[id] = (event.target as HTMLInputElement).files?.[0] || null;
};
const chooseCreationFile = (event: Event) => {
  creationFile.value = (event.target as HTMLInputElement).files?.[0] || null;
};

const uploadResourceFile = async (item: Item) => {
  const id = itemId(item);
  const file = id ? selectedFiles[id] : null;
  if (!id || !file) return;
  const body = new FormData();
  body.append("file", file);
  const resource =
    activeTab.value === "materials" ? "materials" : "assignments";
  try {
    await request(`/teacher/${resource}/${id}/files`, { method: "POST", body });
    successMessage.value = "File berhasil diupload.";
    selectedFiles[id] = null;
    await loadItems();
  } catch (error: any) {
    errorMessage.value = error?.data?.message || "File belum dapat diupload.";
  }
};
const resourceFiles = (item: Item) =>
  pick(
    item,
    "Files",
    "files",
    "MaterialFiles",
    "material_files",
    "AssignmentFiles",
    "assignment_files",
    "attachments",
  ) || [];
const downloadResourceFile = async (
  file: any,
  kind: "material" | "assignment",
) => {
  try {
    const studentPath =
      kind === "material"
        ? `/student/material-files/${optionId(file)}/download`
        : `/student/assignment-files/${optionId(file)}/download`;
    const teacherPath =
      kind === "material"
        ? `/teacher/material-files/${optionId(file)}/download`
        : `/teacher/assignment-files/${optionId(file)}/download`;
    const path = isTeacher.value ? teacherPath : studentPath;
    const blob = await request<Blob>(path, { responseType: "blob" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download =
      pick(file, "OriginalName", "original_name", "FileName", "file_name") ||
      "lampiran";
    anchor.click();
    URL.revokeObjectURL(url);
  } catch (error: any) {
    setApiError(error, "Lampiran belum dapat diunduh.");
  }
};
const deleteResourceFile = async (
  item: Item,
  file: any,
  kind: "material" | "assignment",
) => {
  const resource = kind === "material" ? "materials" : "assignments";
  if (!confirm("Hapus lampiran ini?")) return;
  await action(
    "DELETE",
    `/teacher/${resource}/${itemId(item)}/files/${optionId(file)}`,
    "Lampiran berhasil dihapus.",
  );
};
const loadSubmissions = async (assignment: Item) => {
  try {
    const response = await request<{ data?: any[] }>(
      `/teacher/assignments/${itemId(assignment)}/submissions`,
    );
    submissionRows.value = response.data || [];
    submissionAssignmentTitle.value = itemTitle(assignment);
    submissionAssignment.value = assignment;
    selectedSubmission.value = null;
  } catch (error: any) {
    setApiError(error, "Rekap pengumpulan belum dapat dimuat.");
  }
};
const openSubmission = async (row: any) => {
  try {
    const response = await request<{ data?: any }>(
      `/teacher/submissions/${row.id}`,
    );
    selectedSubmission.value = response.data || row;
    gradeForm.value = {
      score: String(
        response.data?.score ?? response.data?.Score ?? row.score ?? "",
      ),
      feedback: response.data?.feedback ?? response.data?.Feedback ?? "",
    };
  } catch (error: any) {
    setApiError(error, "Detail pengumpulan belum dapat dimuat.");
  }
};
const downloadSubmissionFile = async (file: any) => {
  try {
    const blob = await request<Blob>(
      `/teacher/submission-files/${optionId(file)}/download`,
      { responseType: "blob" },
    );
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download =
      pick(file, "OriginalName", "original_name") || "submission";
    anchor.click();
    URL.revokeObjectURL(url);
  } catch (error: any) {
    setApiError(error, "File submission belum dapat diunduh.");
  }
};
const gradeSubmission = async () => {
  if (!selectedSubmission.value || gradeForm.value.score === "") return;
  grading.value = true;
  try {
    await request(
      `/teacher/submissions/${optionId(selectedSubmission.value)}/grade`,
      {
        method: "PATCH",
        body: {
          score: Number(gradeForm.value.score),
          feedback: gradeForm.value.feedback || null,
        },
      },
    );
    successMessage.value = "Nilai berhasil disimpan.";
    await openSubmission(selectedSubmission.value);
    await loadSubmissions({
      id:
        selectedSubmission.value.Assignment?.ID ||
        selectedSubmission.value.Assignment?.id ||
        selectedSubmission.value.assignment_id,
      title: submissionAssignmentTitle.value,
    });
  } catch (error: any) {
    setApiError(error, "Nilai belum dapat disimpan.");
  } finally {
    grading.value = false;
  }
};
const exportSubmissionsCsv = () => {
  const escapeCsv = (value: any) =>
    `"${String(value ?? "").replaceAll('"', '""')}"`;
  const assignment = submissionAssignment.value || {};
  const subject = optionName(
    pick(assignment, "ClassSubject.Subject", "class_subject.subject"),
  );
  const className = optionName(
    pick(assignment, "ClassSubject.Class", "class_subject.class"),
  );
  const maxScore = pick(assignment, "max_score", "MaxScore") || "";
  const rows = [
    [
      "No",
      "NIS",
      "Nama Lengkap",
      "Mata Pelajaran",
      "Kelas",
      "Nama Tugas",
      "Status Pengumpulan",
      "Waktu Pengumpulan",
      "Nilai",
      "Nilai Maksimal",
    ],
    ...submissionRows.value.map((row, index) => [
      index + 1,
      row.nis || "",
      row.student_name || "",
      subject,
      className,
      submissionAssignmentTitle.value,
      row.status || "Belum dikumpulkan",
      row.submitted_at
        ? new Date(row.submitted_at).toLocaleString("id-ID")
        : "-",
      row.score ?? "",
      maxScore,
    ]),
  ];
  const csv =
    "\uFEFF" + rows.map((row) => row.map(escapeCsv).join(";")).join("\r\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `rekap-${submissionAssignmentTitle.value || "tugas"}.csv`;
  anchor.click();
  URL.revokeObjectURL(url);
};
const printSubmissions = () => {
  document.body.classList.add("printing-report");
  window.setTimeout(() => {
    window.print();
    document.body.classList.remove("printing-report");
  }, 0);
};

const checkIn = async (item: Item) => {
  const id = itemId(item);
  if (!id) return;
  try {
    await request(`/student/attendance-sessions/${encodeURIComponent(id)}/check-in`, {
      method: "POST",
    });
    item.checked_in = true;
    item.attendance_status = "present";
    successMessage.value = "Presensi berhasil dicatat.";
    if (isStudent.value) {
      request<{ data?: any }>("/student/attendance-statistics")
        .then((res) => {
          studentAttendanceStats.value = res.data || null;
        })
        .catch(() => {});
    }
    await loadItems();
  } catch (error: any) {
    setApiError(error, "Gagal melakukan presensi.");
  }
};

const submitAssignment = async (item: Item) => {
  const id = assignmentId(item);
  const file = id ? selectedFiles[id] : null;
  if (!id || !file) return;
  const body = new FormData();
  body.append("file", file);
  try {
    await request(`/student/assignments/${encodeURIComponent(id)}/submit`, {
      method: "POST",
      body,
    });
    successMessage.value = "Tugas berhasil dikumpulkan.";
    selectedFiles[id] = null;
    await loadItems();
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Tugas belum dapat dikumpulkan.";
  }
};

const itemTitle = (item: Item) =>
  pick(
    item,
    "title",
    "Title",
    "name",
    "Name",
    "Student.FullName",
    "Teacher.FullName",
    "full_name",
    "FullName",
    "username",
    "Username",
  ) || "Data tanpa nama";
const itemMeta = (item: Item) => {
  if (activeTab.value === "attendance") {
    const d = item.attendance_date || pick(item, "AttendanceDate", "attendance_date") || "";
    const st = item.start_time || pick(item, "StartTime", "start_time") || "";
    const et = item.end_time || pick(item, "EndTime", "end_time") || "";
    const timeStr = st ? ` · ${st.slice(11, 16) || st} - ${et.slice(11, 16) || et}` : "";
    const statusStr = item.status === "open" ? "Dibuka" : "Ditutup";
    return `${statusStr} · ${d || "Hari ini"}${timeStr}`;
  }
  if (
    isStudent.value &&
    ["materials", "assignments", "quizzes"].includes(activeTab.value)
  ) {
    return activeConfig.value?.label || "Aktivitas";
  }
  return (
    pick(
      item,
      "role.name",
      "role.Name",
      "Role.Name",
      "status",
      "Status",
      "email",
      "Email",
      "username",
      "Username",
    ) ||
    (item.attendance_date
      ? `${item.attendance_date} ${item.start_time || ""}`
      : "Aktif")
  );
};
const itemDescriptionHtml = (item: Item) =>
  pick(item, "description", "Description") ||
  (item.deadline
    ? `Tenggat ${new Date(item.deadline).toLocaleDateString("id-ID", { day: "numeric", month: "long" })}`
    : "Belum ada deskripsi tambahan.");
const itemDescription = (item: Item) =>
  itemDescriptionHtml(item).replace(/<[^>]*>/g, " ").replace(/\s+/g, " ").trim();
const itemStatus = (item: Item) =>
  String(pick(item, "status", "Status") || "draft").toLowerCase();
const isDraft = (item: Item) => itemStatus(item) === "draft";
const quizAttemptsUsed = (item: Item) =>
  Number(pick(item, "attempts_used", "AttemptsUsed") || 0);
const quizMaxAttempts = (item: Item) =>
  Number(pick(item, "max_attempts", "MaxAttempts") || 1);
const quizAttemptsExhausted = (item: Item) =>
  activeTab.value === "quizzes" &&
  isStudent.value &&
  quizAttemptsUsed(item) >= quizMaxAttempts(item);
const resourceContext = (item: Item) => {
  const classSubject = pick(item, "ClassSubject", "class_subject");
  return `${optionName(pick(classSubject, "Subject", "subject"))} · ${optionName(pick(classSubject, "Class", "class"))} · Guru: ${optionName(pick(item, "Teacher", "teacher"))}`;
};
const materialGroups = computed(() => {
  if (activeTab.value !== "materials" || !isStudent.value) return [];
  const groups = new Map<
    string,
    {
      key: string;
      subject: string;
      className: string;
      teacher: string;
      items: Item[];
    }
  >();
  for (const item of displayedItems.value) {
    const subject = optionName(
      pick(item, "ClassSubject.Subject", "class_subject.subject"),
    );
    const className = optionName(
      pick(item, "ClassSubject.Class", "class_subject.class"),
    );
    const teacher = optionName(
      pick(
        item,
        "Teacher",
        "teacher",
        "ClassSubject.Teacher",
        "class_subject.teacher",
      ),
    );
    const key = `${subject}|${className}|${teacher}`;
    if (!groups.has(key))
      groups.set(key, { key, subject, className, teacher, items: [] });
    groups.get(key)!.items.push(item);
  }
  return [...groups.values()];
});
const detailRole = (item: any) =>
  pick(item, "role.name", "role.Name", "Role.Name", "Role.name") || "-";
const detailClasses = (item: any) =>
  pick(item, "Student.Classes", "student.classes", "Classes", "classes") || [];
const classAdminLink = (item: any) =>
  `/admin?class_id=${encodeURIComponent(optionId(item) || "")}`;
const studentDetailLink = (item: any) =>
  `/learning?tab=students&student_id=${encodeURIComponent(optionId(item) || "")}`;
const detailTeachers = (item: any) => pick(item, "Teachers", "teachers") || [];
const detailStudents = (item: any) => pick(item, "Students", "students") || [];
const detailMajor = (item: any) => pick(item, "Major", "major");
const detailAcademicYear = (item: any) =>
  pick(item, "AcademicYear", "academic_year");
const detailPhone = (item: any) =>
  pick(
    item,
    "Teacher.Phone",
    "Student.Phone",
    "teacher.phone",
    "student.phone",
  ) || "-";
const teacherAssignments = (item: any) =>
  pick(item, "Teacher.ClassSubjects", "teacher.class_subjects") || [];
const subjectClassAssignments = (item: any) =>
  pick(item, "ClassSubjects", "class_subjects") || [];
const classSubjectAssignments = (item: any) =>
  pick(item, "ClassSubjects", "class_subjects") || [];

const copiedField = ref("");
const copyText = (text: string, label: string) => {
  if (!text || text === "-") return;
  navigator.clipboard.writeText(text);
  copiedField.value = label;
  setTimeout(() => {
    if (copiedField.value === label) copiedField.value = "";
  }, 2000);
};

const getInitials = (name?: string) => {
  if (!name) return "U";
  const words = String(name).trim().split(/\s+/);
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
};

const formatGender = (gender?: string) => {
  const g = String(gender || "").toLowerCase().trim();
  if (g === "l" || g === "laki-laki" || g === "male") return "Laki-laki";
  if (g === "p" || g === "perempuan" || g === "female") return "Perempuan";
  return gender || "-";
};

const closeDetailModal = () => {
  selectedDetail.value = null;
};

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && selectedDetail.value) {
    closeDetailModal();
  }
};

watch(activeTab, () => {
  closeForm();
  detailQuery.value = "";
  if (activeTab.value !== "students") classFilter.value = "";
  if (!sortOptions.value.some((option) => option.value === sortBy.value))
    sortBy.value = "name";
  loadItems();
});
onMounted(() => {
  const requestedTab = String(route.query.tab || "");
  activeTab.value = tabs.value.some((tab) => tab.key === requestedTab)
    ? requestedTab
    : tabs.value[0]?.key || "";
  window.addEventListener("keydown", handleKeydown);
});
onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleKeydown);
});
onMounted(loadReferenceData);
</script>

<template>
  <AppShell
    active="learning"
    eyebrow="School Space"
    :title="title"
  >
    <section class="page-intro">
      <p class="eyebrow">Pusat aktivitas</p>
      <h2>
        {{
          isStudent
            ? "Semua langkah belajarmu di satu tempat."
            : "Kelola aktivitas sekolah tanpa berpindah konteks."
        }}
      </h2>
      <p>
        {{
          isStudent
            ? "Buka materi, kumpulkan tugas, kerjakan kuis, dan catat presensi."
            : "Pilih modul untuk melihat, menambah, dan memperbarui data."
        }}
      </p>
    </section>

    <div class="toolbar">
      <div class="tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-button"
          :class="{ active: activeTab === tab.key }"
          :style="activeTab === tab.key ? { '--tab-color': tabColor(tab.key) } : {}"
          type="button"
          @click="activeTab = tab.key"
        >
          <component :is="tabIcon(tab.key)" :size="15" />
          {{ tab.label }}
        </button>
      </div>
      <button
        v-if="activeConfig?.create"
        class="button button-primary btn-add"
        type="button"
        @click="openCreate"
      >
        <Plus :size="15" />
        Tambah {{ activeConfig.label }}
      </button>
    </div>

    <div v-if="successMessage" class="banner banner-success"><CheckCircle2 :size="16" /> {{ successMessage }}</div>
    <MeetingPanel v-if="activeTab === 'meetings'" />
    <template v-else>
    <div v-if="errorMessage" class="form-error">
      <p>{{ errorMessage }}</p>
      <button
        v-if="errorDetails"
        class="learn-more"
        type="button"
        @click="showErrorDetails = !showErrorDetails"
      >
        {{ showErrorDetails ? "Sembunyikan detail" : "Pelajari lebih lanjut" }}
      </button>
      <div v-if="showErrorDetails && errorDetails" class="error-details">
        <strong>Alasan</strong>
        <p>{{ errorDetails.reason }}</p>
        <p v-if="errorDetails.dependency">
          Jumlah materi terkait: {{ errorDetails.dependency }}
        </p>
        <strong>Langkah berikutnya</strong>
        <ol>
          <li v-for="step in errorDetails.next_steps" :key="step">
            {{ step }}
          </li>
        </ol>
      </div>
    </div>

    <!-- Klik "Lihat detail" langsung di kartu di bawah untuk membuka detail item —
         tidak perlu dropdown pemilih terpisah lagi sekarang search benar-benar menyaring kartunya. -->
    <div v-if="auth.role === 'admin' && items.length" class="detail-picker">
      <input
        v-model="detailQuery"
        type="search"
        placeholder="Cari nama, kode, atau email"
      />
      <select v-if="activeTab === 'students'" v-model="classFilter">
        <option value="">Semua kelas</option>
        <option
          v-for="className in classFilterOptions"
          :key="className"
          :value="className"
        >
          {{ className }}
        </option>
      </select>
      <select v-model="sortBy">
        <option
          v-for="option in sortOptions"
          :key="option.value"
          :value="option.value"
        >
          Sortir {{ option.label.toLowerCase() }}
        </option>
      </select>
      <button
        class="button button-secondary"
        type="button"
        @click="sortDirection = sortDirection === 'asc' ? 'desc' : 'asc'"
      >
        {{ sortDirection === "asc" ? "A-Z" : "Z-A" }}
      </button>
    </div>

    <section v-if="showForm" class="panel form-panel">
      <div class="panel-heading">
        <div>
          <p class="eyebrow">Formulir baru</p>
          <h3>Tambah {{ activeConfig?.label }}</h3>
        </div>
        <button class="close-button" type="button" @click="closeForm">×</button>
      </div>
      <form class="form-grid" @submit.prevent="submitForm">
        <label
          v-if="
            ['materials', 'assignments', 'quizzes', 'attendance'].includes(
              activeTab,
            )
          "
          >Kelas & mata pelajaran
          <select v-model="form.class_subject_id" required>
            <option value="">Pilih kelas dan mata pelajaran</option>
            <option
              v-for="option in classSubjectOptions"
              :key="optionId(option)"
              :value="optionId(option)"
            >
              {{ option.label || classSubjectName(option) }}
            </option>
          </select>
        </label>
        <label v-if="isTeacherForm || isStudentForm"
          >Username<input v-model="form.username" required
        /></label>
        <label v-if="isTeacherForm || isStudentForm"
          >Password<input
            v-model="form.password"
            type="password"
            minlength="6"
            :required="!editingId"
          /><small v-if="editingId"
            >Kosongkan jika password tidak diubah.</small
          ></label
        >
        <label v-if="isTeacherForm || isStudentForm"
          >Nama lengkap<input v-model="form.full_name" required
        /></label>
        <label v-if="isTeacherForm || isStudentForm"
          >Email<input v-model="form.email" type="email"
        /></label>
        <label v-if="isTeacherForm">NIP<input v-model="form.nip" /></label>
        <label v-if="isTeacherForm"
          >Telepon<input v-model="form.phone"
        /></label>
        <label v-if="isStudentForm"
          >NIS<input v-model="form.nis" required
        /></label>
        <label v-if="isStudentForm">NISN<input v-model="form.nisn" /></label>
        <label v-if="isStudentForm"
          >Jenis kelamin
          <select v-model="form.gender">
            <option value="">Pilih</option>
            <option value="L">Laki-laki</option>
            <option value="P">Perempuan</option>
          </select>
        </label>
        <label v-if="isStudentForm" class="full"
          >Alamat<textarea v-model="form.address" rows="2" />
        </label>
        <label v-if="activeTab === 'classes'"
          >Nama kelas<input v-model="form.name" required
        /></label>
        <label v-if="activeTab === 'classes'"
          >Tingkat
          <select v-model.number="form.grade" required>
            <option :value="10">10 (X)</option>
            <option :value="11">11 (XI)</option>
            <option :value="12">12 (XII)</option>
          </select>
        </label>
        <label v-if="activeTab === 'classes'"
          >Jurusan
          <select v-model="form.major_id">
            <option value="">Tanpa jurusan</option>
            <option
              v-for="major in majorOptions"
              :key="optionId(major)"
              :value="optionId(major)"
            >
              {{ major.code || major.Code }} - {{ major.name || major.Name }}
            </option>
          </select>
          <small v-if="!majorOptions.length">Belum ada data jurusan.</small>
        </label>
        <label v-if="activeTab === 'classes'"
          >Tahun akademik
          <select v-model="form.academic_year_id" required>
            <option value="">Pilih tahun akademik</option>
            <option
              v-for="year in academicYearOptions"
              :key="optionId(year)"
              :value="optionId(year)"
            >
              {{ optionName(year) }}
            </option>
          </select>
          <small v-if="!academicYearOptions.length"
            >Belum ada data tahun akademik dari backend.</small
          >
        </label>
        <label v-if="activeTab === 'majors'"
          >Kode jurusan<input v-model="form.code" required
        /></label>
        <label v-if="activeTab === 'majors'"
          >Nama jurusan<input v-model="form.name" required
        /></label>
        <label v-if="activeTab === 'majors'" class="full"
          >Deskripsi<RichTextEditor v-model="form.description" />
        </label>
        <label v-if="activeTab === 'subjects'"
          >Kode<input v-model="form.code" required
        /></label>
        <label v-if="activeTab === 'subjects'"
          >Nama mata pelajaran<input v-model="form.name" required
        /></label>
        <fieldset v-if="activeTab === 'subjects'" class="teacher-picker full">
          <legend>Guru yang dapat mengajar mata pelajaran ini</legend>
          <label
            v-for="teacher in teacherOptions"
            :key="optionId(teacher)"
            class="teacher-option"
          >
            <input
              v-model="form.teacher_ids"
              type="checkbox"
              :value="
                pick(teacher, 'Teacher.id', 'Teacher.ID') || optionId(teacher)
              "
            />
            <span>{{
              pick(
                teacher,
                "Teacher.FullName",
                "full_name",
                "FullName",
                "username",
                "Username",
              )
            }}</span>
          </label>
          <small v-if="!teacherOptions.length"
            >Belum ada guru yang tersedia.</small
          >
        </fieldset>
        <label
          v-if="
            ['materials', 'assignments', 'quizzes', 'attendance'].includes(
              activeTab,
            )
          "
          >Judul<input v-model="form.title" required
        /></label>
        <label
          v-if="
            ['materials', 'assignments', 'quizzes', 'attendance'].includes(
              activeTab,
            )
          "
          class="full"
          >Deskripsi<RichTextEditor v-model="form.description" />
        </label>
        <label v-if="activeTab === 'assignments'"
          >Tenggat<input
            v-model="form.deadline"
            type="datetime-local"
            :min="minDeadline"
            required
        /></label>
        <label v-if="activeTab === 'assignments'"
          >Nilai maksimal<input
            v-model.number="form.max_score"
            type="number"
            min="1"
            max="100"
            required
        /></label>
        <label v-if="activeTab === 'quizzes'"
          >Durasi (menit)<input
            v-model.number="form.duration"
            type="number"
            min="1"
            max="300"
            required
        /></label>
        <label v-if="activeTab === 'quizzes'"
          >Maksimal percobaan<input
            v-model.number="form.max_attempts"
            type="number"
            min="1"
            max="10"
            required
        /></label>
        <label v-if="activeTab === 'attendance'"
          >Tanggal<input v-model="form.attendance_date" type="date" required
        /></label>
        <label v-if="activeTab === 'attendance'"
          >Mulai<input v-model="form.start_time" type="time" required
        /></label>
        <label v-if="activeTab === 'attendance'"
          >Selesai<input v-model="form.end_time" type="time" required
        /></label>
        <label v-if="activeTab === 'materials'"
          >URL eksternal<input v-model="form.external_url" type="url"
        /></label>
        <label v-if="activeTab === 'materials'"
          >URL video<input v-model="form.video_url" type="url"
        /></label>
        <label
          v-if="['materials', 'assignments'].includes(activeTab) && !editingId"
          class="full"
          >Lampiran (opsional)<input
            type="file"
            @change="chooseCreationFile"
          /><small v-if="creationFile">{{ creationFile.name }}</small>
        </label>
        <div class="form-actions full">
          <button
            class="button button-secondary"
            type="button"
            @click="closeForm"
          >
            Batal
          </button>
          <button
            class="button button-primary"
            type="submit"
            :disabled="saving"
          >
            {{ saving ? "Menyimpan..." : editingId ? "Perbarui" : "Simpan" }}
          </button>
        </div>
      </form>
    </section>

    <!-- Modal Dialog Lihat Detail -->
    <transition name="fade-slide">
      <div
        v-if="selectedDetail"
        class="modal-backdrop detail-modal-backdrop"
        @click.self="closeDetailModal"
      >
        <div class="modal-card detail-modal-container">
          <!-- Modal Header -->
          <div class="detail-modal-header">
            <div class="detail-header-profile">
              <div
                class="detail-avatar"
                :class="`avatar-${activeTab}`"
              >
                {{
                  getInitials(
                    pick(
                      selectedDetail,
                      "Teacher.FullName",
                      "Student.FullName",
                      "Name",
                      "name",
                      "Title",
                      "title",
                      "username",
                      "Username",
                    ),
                  )
                }}
              </div>
              <div class="detail-header-info">
                <div class="detail-badge-row">
                  <span class="detail-type-pill" :class="`pill-${activeTab}`">
                    <component :is="tabIcon(activeTab)" :size="12" />
                    {{
                      activeTab === "teachers"
                        ? "Guru Pengajar"
                        : activeTab === "students"
                          ? "Siswa"
                          : activeConfig?.label || "Detail"
                    }}
                  </span>
                  <span
                    v-if="activeTab === 'students' && detailClasses(selectedDetail).length"
                    class="detail-class-pill"
                  >
                    {{ optionName(detailClasses(selectedDetail)[0]) }}
                  </span>
                  <span
                    v-else-if="activeTab === 'students'"
                    class="detail-unassigned-pill"
                  >
                    Belum masuk kelas
                  </span>
                </div>
                <h3 class="detail-modal-title">
                  {{
                    pick(
                      selectedDetail,
                      "Teacher.FullName",
                      "Student.FullName",
                    ) || itemTitle(selectedDetail)
                  }}
                </h3>
                <p class="detail-modal-subtitle">
                  <span v-if="pick(selectedDetail, 'username', 'Username')">
                    @{{ pick(selectedDetail, "username", "Username") }}
                  </span>
                  <span
                    v-if="
                      pick(selectedDetail, 'username', 'Username') &&
                      pick(selectedDetail, 'email', 'Email')
                    "
                    class="email-sep"
                    >·</span
                  >
                  <span v-if="pick(selectedDetail, 'email', 'Email')">
                    {{ pick(selectedDetail, "email", "Email") }}
                  </span>
                </p>
              </div>
            </div>
            <button
              class="close-button"
              type="button"
              aria-label="Tutup detail"
              @click="closeDetailModal"
            >
              ×
            </button>
          </div>

          <!-- Modal Scrollable Content -->
          <div class="detail-modal-body">
            <!-- SPECIFIC VIEW: STUDENTS -->
            <template v-if="activeTab === 'students'">
              <!-- Card 1: Data Akademik & Kelas -->
              <div class="detail-card-group">
                <div class="group-header">
                  <GraduationCap :size="16" class="group-icon" />
                  <h4>Kelas & Status Akademik</h4>
                </div>
                <div class="group-content">
                  <div
                    v-if="detailClasses(selectedDetail).length"
                    class="student-class-highlight"
                  >
                    <div class="class-info-left">
                      <span class="class-main-name">
                        {{ optionName(detailClasses(selectedDetail)[0]) }}
                      </span>
                      <span class="class-sub-meta">
                        {{
                          detailClasses(selectedDetail)[0]?.grade
                            ? `Tingkat ${detailClasses(selectedDetail)[0].grade}`
                            : ""
                        }}
                        {{
                          detailClasses(selectedDetail)[0]?.major?.name
                            ? `· ${detailClasses(selectedDetail)[0].major.name}`
                            : ""
                        }}
                      </span>
                    </div>
                    <NuxtLink
                      class="btn-open-class"
                      :to="classAdminLink(detailClasses(selectedDetail)[0])"
                      @click="closeDetailModal"
                    >
                      Buka di Administrasi <ExternalLink :size="12" />
                    </NuxtLink>
                  </div>
                  <div v-else class="student-no-class-alert">
                    <AlertCircle :size="16" class="alert-icon" />
                    <div>
                      <strong>Belum Terhubung ke Kelas</strong>
                      <p>Siswa ini belum dimasukkan ke dalam rombongan belajar / kelas manapun.</p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Card 2: Identitas Personal Siswa -->
              <div class="detail-card-group">
                <div class="group-header">
                  <User :size="16" class="group-icon" />
                  <h4>Identitas Pribadi Siswa</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">NIS (Nomor Induk Siswa)</span>
                    <div class="info-value-row">
                      <span class="info-value font-mono">{{ pick(selectedDetail, "Student.NIS") || "-" }}</span>
                      <button
                        v-if="pick(selectedDetail, 'Student.NIS')"
                        type="button"
                        class="btn-copy"
                        title="Salin NIS"
                        @click="copyText(pick(selectedDetail, 'Student.NIS'), 'nis')"
                      >
                        <component :is="copiedField === 'nis' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">NISN</span>
                    <div class="info-value-row">
                      <span class="info-value font-mono">{{ pick(selectedDetail, "Student.NISN") || "-" }}</span>
                      <button
                        v-if="pick(selectedDetail, 'Student.NISN')"
                        type="button"
                        class="btn-copy"
                        title="Salin NISN"
                        @click="copyText(pick(selectedDetail, 'Student.NISN'), 'nisn')"
                      >
                        <component :is="copiedField === 'nisn' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Jenis Kelamin</span>
                    <span class="info-value">
                      {{ formatGender(pick(selectedDetail, "Student.Gender")) }}
                    </span>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Peran Pengguna</span>
                    <span class="info-value">{{ detailRole(selectedDetail) }}</span>
                  </div>

                  <div class="info-cell full-width">
                    <span class="info-label">Alamat Lengkap</span>
                    <span class="info-value">
                      <MapPin :size="13" class="inline-icon" />
                      {{ pick(selectedDetail, "Student.Address") || "Belum diisi" }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Card 3: Kontak & Akun -->
              <div class="detail-card-group">
                <div class="group-header">
                  <Phone :size="16" class="group-icon" />
                  <h4>Kontak & Akun Login</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Nomor Telepon / WhatsApp</span>
                    <div class="info-value-row">
                      <span class="info-value">{{ detailPhone(selectedDetail) }}</span>
                      <button
                        v-if="detailPhone(selectedDetail) !== '-'"
                        type="button"
                        class="btn-copy"
                        title="Salin Telepon"
                        @click="copyText(detailPhone(selectedDetail), 'phone')"
                      >
                        <component :is="copiedField === 'phone' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Alamat Email</span>
                    <div class="info-value-row">
                      <span class="info-value">{{ pick(selectedDetail, "email", "Email") || "-" }}</span>
                      <button
                        v-if="pick(selectedDetail, 'email', 'Email')"
                        type="button"
                        class="btn-copy"
                        title="Salin Email"
                        @click="copyText(pick(selectedDetail, 'email', 'Email'), 'email')"
                      >
                        <component :is="copiedField === 'email' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Username Login</span>
                    <span class="info-value">@{{ pick(selectedDetail, "username", "Username") || "-" }}</span>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Status Akun</span>
                    <span class="info-value status-active-text">● Aktif</span>
                  </div>
                </div>
              </div>
            </template>

            <!-- SPECIFIC VIEW: TEACHERS -->
            <template v-else-if="activeTab === 'teachers'">
              <!-- Card 1: Data Pengajar -->
              <div class="detail-card-group">
                <div class="group-header">
                  <Shield :size="16" class="group-icon" />
                  <h4>Data Pengajar & Akun</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">NIP (Nomor Induk Pegawai)</span>
                    <div class="info-value-row">
                      <span class="info-value font-mono">{{ pick(selectedDetail, "Teacher.NIP") || "-" }}</span>
                      <button
                        v-if="pick(selectedDetail, 'Teacher.NIP')"
                        type="button"
                        class="btn-copy"
                        title="Salin NIP"
                        @click="copyText(pick(selectedDetail, 'Teacher.NIP'), 'nip')"
                      >
                        <component :is="copiedField === 'nip' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Peran</span>
                    <span class="info-value">{{ detailRole(selectedDetail) }}</span>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Telepon / WhatsApp</span>
                    <div class="info-value-row">
                      <span class="info-value">{{ detailPhone(selectedDetail) }}</span>
                      <button
                        v-if="detailPhone(selectedDetail) !== '-'"
                        type="button"
                        class="btn-copy"
                        title="Salin Telepon"
                        @click="copyText(detailPhone(selectedDetail), 'phone')"
                      >
                        <component :is="copiedField === 'phone' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell">
                    <span class="info-label">Alamat Email</span>
                    <div class="info-value-row">
                      <span class="info-value">{{ pick(selectedDetail, "email", "Email") || "-" }}</span>
                      <button
                        v-if="pick(selectedDetail, 'email', 'Email')"
                        type="button"
                        class="btn-copy"
                        title="Salin Email"
                        @click="copyText(pick(selectedDetail, 'email', 'Email'), 'email')"
                      >
                        <component :is="copiedField === 'email' ? CheckCheck : Copy" :size="13" />
                      </button>
                    </div>
                  </div>

                  <div class="info-cell full-width">
                    <span class="info-label">Username Login</span>
                    <span class="info-value">@{{ pick(selectedDetail, "username", "Username") || "-" }}</span>
                  </div>
                </div>
              </div>

              <!-- Card 2: Penugasan Mengajar -->
              <div class="detail-card-group">
                <div class="group-header">
                  <BookOpen :size="16" class="group-icon" />
                  <h4>Mata Pelajaran & Kelas yang Diampu ({{ teacherAssignments(selectedDetail).length }})</h4>
                </div>
                <div class="group-content">
                  <div
                    v-if="teacherAssignments(selectedDetail).length"
                    class="assignment-cards-list"
                  >
                    <div
                      v-for="assignment in teacherAssignments(selectedDetail)"
                      :key="optionId(assignment)"
                      class="assignment-mini-card"
                    >
                      <div class="assignment-icon-wrap">
                        <BookMarked :size="15" />
                      </div>
                      <div class="assignment-content">
                        <strong>{{ optionName(assignment.Subject) }}</strong>
                        <span>Kelas: {{ optionName(assignment.Class) }}</span>
                      </div>
                    </div>
                  </div>
                  <div v-else class="empty-assignment-box">
                    <AlertCircle :size="16" class="alert-icon" />
                    <p>Guru ini belum ditugaskan untuk mengajar mata pelajaran pada kelas manapun.</p>
                  </div>
                </div>
              </div>
            </template>

            <!-- SPECIFIC VIEW: CLASSES -->
            <template v-else-if="activeTab === 'classes'">
              <div class="detail-card-group">
                <div class="group-header">
                  <Layers :size="16" class="group-icon" />
                  <h4>Informasi Kelas</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Nama Kelas</span>
                    <span class="info-value">{{ pick(selectedDetail, "Name", "name") || "-" }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Tingkat (Grade)</span>
                    <span class="info-value">Tingkat {{ pick(selectedDetail, "Grade", "grade") || "-" }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Jurusan</span>
                    <span class="info-value">
                      {{ detailMajor(selectedDetail) ? optionName(detailMajor(selectedDetail)) : "Tanpa jurusan" }}
                    </span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Tahun Akademik</span>
                    <span class="info-value">
                      {{ detailAcademicYear(selectedDetail) ? optionName(detailAcademicYear(selectedDetail)) : "-" }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Siswa dalam kelas -->
              <div class="detail-card-group">
                <div class="group-header">
                  <Users :size="16" class="group-icon" />
                  <h4>Siswa Terdaftar ({{ detailStudents(selectedDetail).length }})</h4>
                </div>
                <div v-if="detailStudents(selectedDetail).length" class="chip-grid-wrap">
                  <span
                    v-for="student in detailStudents(selectedDetail)"
                    :key="optionId(student)"
                    class="person-badge-chip"
                  >
                    {{ optionName(student) }}
                  </span>
                </div>
                <p v-else class="empty-hint-text">Belum ada siswa yang terdaftar di kelas ini.</p>
              </div>

              <!-- Mapel dalam kelas -->
              <div class="detail-card-group">
                <div class="group-header">
                  <BookOpen :size="16" class="group-icon" />
                  <h4>Mata Pelajaran & Guru Pengajar ({{ classSubjectAssignments(selectedDetail).length }})</h4>
                </div>
                <div v-if="classSubjectAssignments(selectedDetail).length" class="assignment-cards-list">
                  <div
                    v-for="assignment in classSubjectAssignments(selectedDetail)"
                    :key="optionId(assignment)"
                    class="assignment-mini-card"
                  >
                    <div class="assignment-icon-wrap"><BookMarked :size="15" /></div>
                    <div class="assignment-content">
                      <strong>{{ optionName(assignment.Subject) }}</strong>
                      <span>Guru: {{ optionName(assignment.Teacher) || "Belum ditentukan" }}</span>
                    </div>
                  </div>
                </div>
                <p v-else class="empty-hint-text">Belum ada mata pelajaran yang ditugaskan di kelas ini.</p>
              </div>
            </template>

            <!-- SPECIFIC VIEW: SUBJECTS -->
            <template v-else-if="activeTab === 'subjects'">
              <div class="detail-card-group">
                <div class="group-header">
                  <BookMarked :size="16" class="group-icon" />
                  <h4>Informasi Mata Pelajaran</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Kode Mata Pelajaran</span>
                    <span class="info-value font-mono">{{ pick(selectedDetail, "Code", "code") || "-" }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Nama Mata Pelajaran</span>
                    <span class="info-value">{{ pick(selectedDetail, "Name", "name") || "-" }}</span>
                  </div>
                  <div class="info-cell full-width">
                    <span class="info-label">Deskripsi</span>
                    <span class="info-value">{{ pick(selectedDetail, "Description", "description") || "Tidak ada deskripsi." }}</span>
                  </div>
                  <div class="info-cell full-width">
                    <span class="info-label">Guru Pengampu</span>
                    <span class="info-value">
                      {{ detailTeachers(selectedDetail).map((item: any) => optionName(item)).join(", ") || "Belum ada guru pengampu" }}
                    </span>
                  </div>
                </div>
              </div>
            </template>

            <!-- SPECIFIC VIEW: MAJORS -->
            <template v-else-if="activeTab === 'majors'">
              <div class="detail-card-group">
                <div class="group-header">
                  <LayoutGrid :size="16" class="group-icon" />
                  <h4>Informasi Jurusan</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Kode Jurusan</span>
                    <span class="info-value font-mono">{{ pick(selectedDetail, "Code", "code") || "-" }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Nama Jurusan</span>
                    <span class="info-value">{{ pick(selectedDetail, "Name", "name") || "-" }}</span>
                  </div>
                  <div class="info-cell full-width">
                    <span class="info-label">Deskripsi</span>
                    <span class="info-value">{{ pick(selectedDetail, "Description", "description") || "Tidak ada deskripsi." }}</span>
                  </div>
                </div>
              </div>
            </template>

            <!-- SPECIFIC VIEW: MATERIALS -->
            <template v-else-if="activeTab === 'materials'">
              <div class="detail-card-group">
                <div class="group-header">
                  <BookOpen :size="16" class="group-icon" />
                  <h4>Informasi Materi Pembelajaran</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Judul Materi</span>
                    <span class="info-value">{{ itemTitle(selectedDetail) }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Mata Pelajaran & Kelas</span>
                    <span class="info-value">
                      {{ optionName(pick(selectedDetail, "ClassSubject.Subject", "class_subject.subject")) }} ·
                      {{ optionName(pick(selectedDetail, "ClassSubject.Class", "class_subject.class")) }}
                    </span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Guru Pengajar</span>
                    <span class="info-value">{{ optionName(pick(selectedDetail, "Teacher", "teacher")) }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Status Publikasi</span>
                    <span class="info-value">{{ itemStatus(selectedDetail) }}</span>
                  </div>
                  <div class="info-cell full-width">
                    <span class="info-label">Deskripsi / Isi Materi</span>
                    <RichTextContent :content="itemDescriptionHtml(selectedDetail)" />
                  </div>
                </div>
              </div>

              <!-- Lampiran -->
              <div v-if="resourceFiles(selectedDetail).length" class="detail-card-group">
                <div class="group-header">
                  <Paperclip :size="16" class="group-icon" />
                  <h4>Berkas Lampiran</h4>
                </div>
                <div class="files-attachment-list">
                  <button
                    v-for="file in resourceFiles(selectedDetail)"
                    :key="optionId(file)"
                    type="button"
                    class="file-download-chip"
                    @click="downloadResourceFile(file, 'material')"
                  >
                    <Download :size="13" />
                    {{ pick(file, "OriginalName", "original_name") || "lampiran" }}
                  </button>
                </div>
              </div>
            </template>

            <!-- SPECIFIC VIEW: ASSIGNMENTS -->
            <template v-else-if="activeTab === 'assignments'">
              <div class="detail-card-group">
                <div class="group-header">
                  <ClipboardList :size="16" class="group-icon" />
                  <h4>Informasi Tugas</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Judul Tugas</span>
                    <span class="info-value">{{ itemTitle(selectedDetail) }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Batas Waktu (Deadline)</span>
                    <span class="info-value font-semibold">
                      {{ pick(selectedDetail, "deadline", "Deadline") ? new Date(pick(selectedDetail, "deadline", "Deadline")).toLocaleString("id-ID") : "-" }}
                    </span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Nilai Maksimal</span>
                    <span class="info-value">{{ pick(selectedDetail, "max_score", "MaxScore") || 100 }} poin</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Mata Pelajaran & Kelas</span>
                    <span class="info-value">
                      {{ optionName(pick(selectedDetail, "ClassSubject.Subject", "class_subject.subject")) }} ·
                      {{ optionName(pick(selectedDetail, "ClassSubject.Class", "class_subject.class")) }}
                    </span>
                  </div>
                  <div class="info-cell full-width">
                    <span class="info-label">Petunjuk & Deskripsi Tugas</span>
                    <RichTextContent :content="itemDescriptionHtml(selectedDetail)" />
                  </div>
                </div>
              </div>

              <!-- Student Submission actions if student -->
              <div v-if="isStudent" class="detail-card-group">
                <div class="group-header">
                  <Activity :size="16" class="group-icon" />
                  <h4>Status Pengumpulan Anda</h4>
                </div>
                <div class="submission-box-highlight">
                  <div class="submission-status-badge">
                    {{ studentSubmission?.Status || studentSubmission?.status || "Belum dikumpulkan" }}
                  </div>
                  <p v-if="studentSubmission?.SubmittedAt || studentSubmission?.submitted_at" class="submission-time-text">
                    Dikumpulkan pada: {{ new Date(studentSubmission.SubmittedAt || studentSubmission.submitted_at).toLocaleString("id-ID") }}
                  </p>
                  <p v-if="studentSubmission?.Score !== undefined || studentSubmission?.score !== undefined" class="submission-score-text">
                    Nilai: <strong>{{ studentSubmission.Score ?? studentSubmission.score }}</strong>
                  </p>
                  <div v-if="!studentSubmission" class="upload-action-row">
                    <input type="file" @change="chooseFile(selectedDetail, $event)" />
                    <button
                      class="button button-primary"
                      type="button"
                      :disabled="!selectedFiles[itemId(selectedDetail)]"
                      @click="submitAssignment(selectedDetail)"
                    >
                      Kumpulkan Tugas Sekarang
                    </button>
                  </div>
                </div>
              </div>
            </template>

            <!-- FALLBACK VIEW -->
            <template v-else>
              <div class="detail-card-group">
                <div class="group-header">
                  <h4>Detail Informasi</h4>
                </div>
                <div class="info-grid-2">
                  <div class="info-cell">
                    <span class="info-label">Nama / Judul</span>
                    <span class="info-value">{{ itemTitle(selectedDetail) }}</span>
                  </div>
                  <div class="info-cell">
                    <span class="info-label">Metadata</span>
                    <span class="info-value">{{ itemMeta(selectedDetail) }}</span>
                  </div>
                  <div class="info-cell full-width">
                    <span class="info-label">Deskripsi</span>
                    <RichTextContent :content="itemDescriptionHtml(selectedDetail)" />
                  </div>
                </div>
              </div>
            </template>
          </div>

          <!-- Modal Footer Actions -->
          <div class="detail-modal-footer">
            <div class="modal-footer-left">
              <transition name="fade-slide">
                <span v-if="copiedField" class="copied-toast">
                  ✓ Berhasil menyalin {{ copiedField.toUpperCase() }}
                </span>
              </transition>
            </div>
            <div class="modal-footer-right">
              <button
                class="button button-secondary"
                type="button"
                @click="closeDetailModal"
              >
                Tutup
              </button>
              <button
                v-if="isTeacher && activeTab === 'assignments'"
                class="button button-secondary"
                type="button"
                @click="loadSubmissions(selectedDetail); closeDetailModal()"
              >
                Lihat Pengumpulan
              </button>
              <button
                v-if="!isStudent"
                class="button button-primary"
                type="button"
                @click="openEdit(selectedDetail); closeDetailModal()"
              >
                Edit Data
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <!-- Both grids now render `displayedItems` so search/filter/sort actually take effect. -->
    <section v-if="auth.role === 'admin' && !loading" class="admin-card-grid">
      <article
        v-for="(item, index) in displayedItems"
        :key="itemId(item) || index"
        class="admin-data-card"
      >
        <div class="card-heading">
          <div class="item-mark" :style="{ background: tabBg(activeTab), color: tabColor(activeTab) }">
            <component :is="tabIcon(activeTab)" :size="18" />
          </div>
          <div>
            <p class="eyebrow">{{ itemMeta(item) }}</p>
            <h3>{{ itemTitle(item) }}</h3>
          </div>
        </div>
        <p>{{ itemDescription(item) }}</p>
        <div class="card-actions">
          <button
            class="button button-secondary"
            type="button"
            @click="openDetail(item)"
          >
            Lihat detail
          </button>
          <button
            class="button button-primary"
            type="button"
            @click="openEdit(item)"
          >
            Edit
          </button>
          <button
            class="button button-danger"
            type="button"
            @click="remove(item)"
          >
            Delete
          </button>
        </div>
      </article>
      <div v-if="!displayedItems.length" class="empty-state">
        Tidak ada data yang cocok dengan pencarian.
      </div>
    </section>

    <div v-if="loading" class="loading-state">
      <span class="loader" /> Memuat data...
    </div>
    <section
      v-else-if="isStudent && activeTab === 'materials'"
      class="material-groups"
    >
      <article
        v-for="group in materialGroups"
        :key="group.key"
        class="material-group panel"
      >
        <div class="material-group-heading">
          <div>
            <p class="eyebrow">{{ group.className }}</p>
            <h3>{{ group.subject }}</h3>
          </div>
          <span class="teacher-badge">Guru: {{ group.teacher }}</span>
        </div>
        <div class="material-group-items">
          <article
            v-for="item in group.items"
            :key="itemId(item)"
            class="learning-item"
          >
            <div class="item-mark" :style="{ background: tabBg('materials'), color: tabColor('materials') }">
              <BookOpen :size="20" />
            </div>
            <div class="item-main">
              <p class="eyebrow">{{ itemMeta(item) }}</p>
              <h3>{{ itemTitle(item) }}</h3>
              <p>{{ itemDescription(item) }}</p>
              <small class="resource-context"
                >{{ group.subject }} · {{ group.className }} · Guru:
                {{ group.teacher }}</small
              >
              <button
                class="button button-secondary resource-open"
                type="button"
                @click="openDetail(item)"
              >
                Buka detail materi
              </button>
              <div v-if="resourceFiles(item).length" class="resource-files">
                <button
                  v-for="file in resourceFiles(item)"
                  :key="optionId(file)"
                  type="button"
                  class="file-link"
                  @click="downloadResourceFile(file, 'material')"
                >
                  Unduh
                  {{
                    pick(file, "OriginalName", "original_name") || "lampiran"
                  }}
                </button>
              </div>
            </div>
          </article>
        </div>
      </article>
      <div v-if="!materialGroups.length" class="empty-state">
        Belum ada materi untuk kelasmu.
      </div>
    </section>
    <section
      v-if="isStudent && activeTab === 'attendance' && studentAttendanceStats"
      class="attendance-stats-bar"
    >
      <div class="stat-mini-card stat-rate">
        <BarChart2 :size="18" class="stat-mini-icon" />
        <span class="stat-mini-val">{{ Math.round(studentAttendanceStats.attendance_percentage || 0) }}%</span>
        <span class="stat-mini-lbl">Kehadiran</span>
      </div>
      <div class="stat-mini-card">
        <Activity :size="18" class="stat-mini-icon" />
        <span class="stat-mini-val">{{ studentAttendanceStats.total_sessions || 0 }}</span>
        <span class="stat-mini-lbl">Total Sesi</span>
      </div>
      <div class="stat-mini-card stat-present">
        <CheckCircle2 :size="18" class="stat-mini-icon" />
        <span class="stat-mini-val">{{ studentAttendanceStats.attended_sessions || 0 }}</span>
        <span class="stat-mini-lbl">Hadir</span>
      </div>
      <div class="stat-mini-card stat-excused">
        <AlertCircle :size="18" class="stat-mini-icon" />
        <span class="stat-mini-val">{{ studentAttendanceStats.excused_sessions || 0 }}</span>
        <span class="stat-mini-lbl">Izin / Sakit</span>
      </div>
      <div class="stat-mini-card stat-absent">
        <XCircle :size="18" class="stat-mini-icon" />
        <span class="stat-mini-val">{{ studentAttendanceStats.absent_sessions || 0 }}</span>
        <span class="stat-mini-lbl">Alpa</span>
      </div>
    </section>
    <section v-if="activeTab !== 'materials' || !isStudent" class="item-list">
      <article
        v-for="(item, index) in displayedItems"
        :key="itemId(item) || index"
        class="learning-item"
      >
        <div class="item-mark" :style="{ background: tabBg(activeTab), color: tabColor(activeTab) }">
          <component :is="tabIcon(activeTab)" :size="20" />
        </div>
        <div class="item-main">
          <p class="eyebrow">{{ itemMeta(item) }}</p>
          <h3>{{ itemTitle(item) }}</h3>
          <p>{{ itemDescription(item) }}</p>
          <button
            v-if="isStudent && activeTab === 'assignments'"
            class="button button-secondary resource-open"
            type="button"
            @click="openDetail(item)"
          >
            Buka detail tugas
          </button>
          <div v-if="activeTab === 'assignments'" class="resource-links">
            <a
              v-if="pick(item, 'video_url', 'VideoURL')"
              :href="pick(item, 'video_url', 'VideoURL')"
              target="_blank"
              rel="noopener"
              >Video</a
            >
            <a
              v-if="pick(item, 'external_url', 'ExternalURL')"
              :href="pick(item, 'external_url', 'ExternalURL')"
              target="_blank"
              rel="noopener"
              >Materi / referensi</a
            >
          </div>
          <div
            v-if="
              resourceFiles(item).length &&
              isStudent &&
              activeTab === 'assignments'
            "
            class="resource-files"
          >
            <button
              v-for="file in resourceFiles(item)"
              :key="optionId(file)"
              type="button"
              class="file-link"
              @click="downloadResourceFile(file, 'assignment')"
            >
              Unduh
              {{ pick(file, "OriginalName", "original_name") || "lampiran" }}
            </button>
          </div>
          <div
            v-if="isStudent && activeTab === 'assignments'"
            class="submission-row"
          >
            <input type="file" @change="chooseFile(item, $event)" />
            <button
              class="button button-primary"
              type="button"
              :disabled="!selectedFiles[itemId(item)]"
              @click="submitAssignment(item)"
            >
              Kumpulkan
            </button>
          </div>
          <div
            v-if="
              isTeacher &&
              ['materials', 'assignments'].includes(activeTab) &&
              isDraft(item)
            "
            class="submission-row"
          >
            <input type="file" @change="chooseFile(item, $event)" />
            <button
              class="button button-secondary"
              type="button"
              :disabled="!selectedFiles[itemId(item)]"
              @click="uploadResourceFile(item)"
            >
              Upload file
            </button>
          </div>
          <div
            v-if="
              isTeacher &&
              ['materials', 'assignments'].includes(activeTab) &&
              resourceFiles(item).length
            "
            class="resource-files"
          >
            <span
              v-for="file in resourceFiles(item)"
              :key="optionId(file)"
              class="file-name"
              >{{
                pick(file, "OriginalName", "original_name") || "lampiran"
              }}</span
            >
            <template v-if="isDraft(item)"
              ><button
                v-for="file in resourceFiles(item)"
                :key="`delete-${optionId(file)}`"
                type="button"
                class="file-delete"
                @click="
                  deleteResourceFile(
                    item,
                    file,
                    activeTab === 'materials' ? 'material' : 'assignment',
                  )
                "
              >
                Hapus
              </button></template
            >
          </div>
        </div>
        <div class="item-actions">
          <button
            v-if="isTeacher && activeTab === 'assignments'"
            class="btn-icon-action"
            type="button"
            @click="openDetail(item)"
          >
            <Eye :size="14" /> Lihat
          </button>
          <NuxtLink
            v-if="isTeacher && activeTab === 'quizzes'"
            class="btn-icon-action"
            :to="`/learning/quiz/${itemId(item)}`"
          ><Brain :size="14" /> Kelola kuis</NuxtLink
          >
          <button
            v-if="isStudent && activeTab === 'quizzes' && !quizAttemptsExhausted(item)"
            class="button button-primary"
            type="button"
            @click="navigateTo(`/learning/quiz/${itemId(item)}`)"
          >
            Mulai kuis
          </button>
          <button
            v-if="isStudent && activeTab === 'quizzes' && quizAttemptsUsed(item) > 0"
            class="button button-secondary"
            type="button"
            @click="navigateTo(`/learning/quiz/${itemId(item)}`)"
          >
            Lihat hasil
          </button>
          <span
            v-else-if="isStudent && activeTab === 'quizzes' && quizAttemptsUsed(item) === 0"
            class="quiz-limit-pill"
          >
            Kesempatan habis
          </span>
          <template v-if="isStudent && activeTab === 'attendance'">
            <span v-if="item.checked_in" class="attendance-pill present">
              <CheckCircle2 :size="13" /> {{ attendanceStatusLabel(item.attendance_status) }}
            </span>
            <button
              v-else-if="item.status === 'open'"
              class="button button-primary btn-checkin"
              type="button"
              @click="checkIn(item)"
            >
              <CalendarCheck :size="14" /> Check-in
            </button>
            <span v-else class="attendance-pill closed">
              <XCircle :size="13" /> Sesi Ditutup
            </span>
          </template>
          <template v-if="isTeacher && activeTab === 'attendance'">
            <NuxtLink
              class="btn-icon-action"
              :to="`/learning/attendance/${itemId(item)}`"
            >
              <Users :size="14" /> Kelola
            </NuxtLink>
            <button
              class="btn-icon-action"
              :class="item.status === 'open' ? 'danger' : 'success'"
              type="button"
              @click="toggleAttendanceSession(item)"
            >
              <component :is="item.status === 'open' ? ToggleRight : ToggleLeft" :size="14" />
              {{ item.status === 'open' ? 'Tutup Sesi' : 'Buka Sesi' }}
            </button>
          </template>
          <button
            v-if="
              isTeacher &&
              ['materials', 'assignments', 'quizzes'].includes(activeTab) &&
              isDraft(item)
            "
            class="btn-icon-action"
            type="button"
            @click="openEdit(item)"
          >
            <Edit2 :size="13" /> Edit
          </button>
          <button
            v-if="
              isTeacher &&
              ['materials', 'assignments', 'quizzes'].includes(activeTab) &&
              isDraft(item)
            "
            class="btn-icon-action success"
            type="button"
            @click="publish(item)"
          >
            <Check :size="13" /> Publikasikan
          </button>
          <button
            v-if="isTeacher && activeTab === 'quizzes' && isDraft(item)"
            class="btn-icon-action"
            type="button"
            @click="closeQuiz(item)"
          >
            <X :size="13" /> Tutup
          </button>
          <button
            v-if="
              (isTeacher &&
                ['materials', 'assignments', 'attendance'].includes(activeTab)) ||
              (isTeacher && activeTab === 'quizzes' && isDraft(item)) ||
              auth.role === 'admin'
            "
            class="btn-icon-action danger"
            type="button"
            title="Hapus"
            @click="remove(item)"
          >
            <Trash2 :size="13" />
          </button>
        </div>
      </article>
      <div v-if="!displayedItems.length" class="empty-state">
        Belum ada data {{ activeConfig?.label.toLowerCase() }}.
      </div>
    </section>
    <section
      v-if="
        isTeacher && activeTab === 'assignments' && submissionAssignmentTitle
      "
      class="panel submission-recap"
    >
      <div class="report-heading">
        <div>
          <p class="report-overline">LAPORAN PENILAIAN</p>
          <h2>Rekap Pengumpulan Tugas</h2>
          <p class="report-meta">
            Tugas: <strong>{{ submissionAssignmentTitle }}</strong>
          </p>
          <p class="report-meta">
            Mata Pelajaran:
            {{
              optionName(
                pick(
                  submissionAssignment,
                  "ClassSubject.Subject",
                  "class_subject.subject",
                ),
              )
            }}
            · Kelas:
            {{
              optionName(
                pick(
                  submissionAssignment,
                  "ClassSubject.Class",
                  "class_subject.class",
                ),
              )
            }}
          </p>
        </div>
        <div class="card-actions recap-actions">
          <button
            class="button button-secondary"
            type="button"
            @click="exportSubmissionsCsv"
          >
            Ekspor CSV</button
          ><button
            class="button button-secondary"
            type="button"
            @click="printSubmissions"
          >
            Cetak / PDF
          </button>
        </div>
      </div>
      <div class="submission-table">
        <div class="submission-table-row submission-table-head">
          <strong>No</strong><strong>NIS</strong><strong>Nama Lengkap</strong
          ><strong>Status Pengumpulan</strong><strong>Waktu Pengumpulan</strong
          ><strong>Nilai</strong>
        </div>
        <div
          v-for="row in submissionRows"
          :key="row.id"
          class="submission-table-row"
          role="button"
          tabindex="0"
          @click="openSubmission(row)"
          @keyup.enter="openSubmission(row)"
        >
          <span>{{ submissionRows.indexOf(row) + 1 }}</span
          ><span>{{ row.nis || "-" }}</span
          ><span>{{ row.student_name }}</span
          ><span>{{ row.status || "Belum dikumpulkan" }}</span
          ><span>{{
            row.submitted_at
              ? new Date(row.submitted_at).toLocaleString("id-ID")
              : "-"
          }}</span
          ><span
            >{{ row.score ?? "-" }} /
            {{
              pick(submissionAssignment, "max_score", "MaxScore") || "-"
            }}</span
          >
        </div>
        <div v-if="!submissionRows.length" class="empty-state">
          Belum ada pengumpulan.
        </div>
      </div>
    </section>
    <section v-if="selectedSubmission" class="panel submission-detail-panel">
      <div class="panel-heading">
        <div>
          <p class="eyebrow">Detail pengumpulan</p>
          <h3>
            {{
              selectedSubmission.StudentName ||
              selectedSubmission.student_name ||
              "Siswa"
            }}
          </h3>
        </div>
        <button
          class="close-button"
          type="button"
          aria-label="Tutup detail"
          @click="selectedSubmission = null"
        >
          ×
        </button>
      </div>
      <div class="detail-grid">
        <div>
          <span>Nama tugas</span
          ><strong>{{
            selectedSubmission.Assignment?.Title ||
            selectedSubmission.assignment?.title ||
            submissionAssignmentTitle
          }}</strong>
        </div>
        <div>
          <span>Status</span
          ><strong>{{
            selectedSubmission.Status || selectedSubmission.status
          }}</strong>
        </div>
        <div>
          <span>Waktu pengumpulan</span
          ><strong>{{
            selectedSubmission.SubmittedAt || selectedSubmission.submitted_at
              ? new Date(
                  selectedSubmission.SubmittedAt ||
                    selectedSubmission.submitted_at,
                ).toLocaleString("id-ID")
              : "-"
          }}</strong>
        </div>
        <div class="detail-wide">
          <span>File siswa</span>
          <div
            v-if="
              (selectedSubmission.Files || selectedSubmission.files || [])
                .length
            "
            class="resource-files"
          >
            <button
              v-for="file in selectedSubmission.Files ||
              selectedSubmission.files"
              :key="optionId(file)"
              type="button"
              class="file-link"
              @click="downloadSubmissionFile(file)"
            >
              Unduh {{ pick(file, "OriginalName", "original_name") }}
            </button>
          </div>
          <strong v-else>Tidak ada file.</strong>
        </div>
        <div>
          <label
            >Nilai<input
              v-model="gradeForm.score"
              type="number"
              min="0"
              :max="
                selectedSubmission.Assignment?.MaxScore ||
                selectedSubmission.assignment?.max_score ||
                100
              "
          /></label>
        </div>
        <div>
          <label
            >Feedback<textarea
              v-model="gradeForm.feedback"
              rows="2"
              placeholder="Feedback untuk siswa"
            />
          </label>
        </div>
      </div>
      <div class="form-actions">
        <button
          class="button button-primary"
          type="button"
          :disabled="grading || gradeForm.score === ''"
          @click="gradeSubmission"
        >
          {{ grading ? "Menyimpan..." : "Simpan nilai" }}
        </button>
      </div>
    </section>
    </template>
  </AppShell>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin: 24px 0 15px;
}
.tabs {
  display: flex;
  gap: 6px;
  flex: 1 1 560px;
  overflow-x: auto;
  padding: 2px;
  scrollbar-width: none;
}
.tabs::-webkit-scrollbar { display: none; }
.tab-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1.5px solid transparent;
  border-radius: 10px;
  padding: 9px 14px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  transition: all 0.18s;
  font-family: 'Manrope', sans-serif;
}
.tab-button:hover {
  background: rgba(94,182,157,0.08);
  color: var(--ink);
}
.tab-button.active {
  border-color: var(--tab-color, #5eb69d);
  background: white;
  color: var(--tab-color, var(--green));
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.detail-picker input {
  flex: 1.6;
  min-width: 180px;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 11px 12px;
  background: white;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.detail-picker input:focus,
.detail-picker select:focus {
  outline: none;
  border-color: #5eb69d;
  box-shadow: 0 0 0 3px rgba(94, 182, 157, 0.16);
}
.detail-picker .button {
  white-space: nowrap;
}
/* Banners */
.banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 11px 16px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
}
.banner-success {
  background: #ecfdf5;
  color: #059669;
  border: 1px solid #a7f3d0;
}
.form-error p {
  margin: 0 0 8px;
}
.learn-more {
  border: 0;
  padding: 0;
  background: transparent;
  color: #8e3d31;
  cursor: pointer;
  text-decoration: underline;
  font: inherit;
  font-size: 11px;
}
.error-details {
  margin-top: 10px;
  border-top: 1px solid #f2cbc3;
  padding-top: 10px;
  line-height: 1.5;
}
.error-details p {
  margin: 4px 0 8px;
}
.error-details ol {
  margin: 4px 0 0 18px;
  padding: 0;
}
.form-panel {
  margin-bottom: 18px;
}
.form-panel .panel-heading {
  margin-bottom: 20px;
}
.close-button {
  border: 0;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  font-size: 22px;
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: start;
  gap: 16px;
}
.form-grid > label,
.form-grid > fieldset {
  min-width: 0;
}
.form-grid > label {
  display: grid;
  align-content: start;
  gap: 7px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 700;
  line-height: 1.35;
}
.form-grid .full {
  grid-column: 1 / -1;
}
.form-grid input,
.form-grid textarea,
.form-grid select {
  width: 100%;
  min-width: 0;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 11px 12px;
  background: white;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.form-grid input:focus,
.form-grid textarea:focus,
.form-grid select:focus {
  outline: none;
  border-color: #5eb69d;
  box-shadow: 0 0 0 3px rgba(94, 182, 157, 0.16);
}
.form-grid textarea {
  resize: vertical;
  min-height: 84px;
}
.form-grid small {
  color: var(--muted);
  font-size: 10px;
  font-weight: 500;
  line-height: 1.45;
}
.teacher-picker {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 12px;
  min-width: 0;
  padding: 14px;
  border: 1px solid #dfe7e1;
  border-radius: 10px;
}
.teacher-picker legend {
  grid-column: 1 / -1;
  padding: 0 4px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 700;
}
.teacher-picker small {
  grid-column: 1 / -1;
}
.teacher-option {
  display: flex !important;
  grid-template-columns: none !important;
  align-items: center;
  gap: 8px !important;
  padding: 8px 9px;
  border: 1px solid #edf2ee;
  border-radius: 8px;
  background: #fbfdfb;
  font-weight: 600 !important;
}
.teacher-option input {
  width: auto;
  flex: 0 0 auto;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.item-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  flex-shrink: 0;
}
/* Icon-style action buttons */
.btn-icon-action {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  border: 1.5px solid #dfe7e1;
  border-radius: 9px;
  padding: 7px 12px;
  background: white;
  color: var(--ink);
  cursor: pointer;
  font: inherit;
  font-size: 12px;
  font-weight: 700;
  font-family: 'Manrope', sans-serif;
  transition: all 0.15s;
  text-decoration: none;
  white-space: nowrap;
}
.btn-icon-action:hover {
  background: var(--paper);
  border-color: #b5cfc0;
}
.btn-icon-action.success {
  border-color: #a7f3d0;
  background: #ecfdf5;
  color: #059669;
}
.btn-icon-action.success:hover { background: #d1fae5; }
.btn-icon-action.danger {
  border-color: #fca5a5;
  background: #fff5f5;
  color: #dc2626;
}
.btn-icon-action.danger:hover { background: #fee2e2; }
/* "Add" button in toolbar */
.btn-add {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
/* Check-in button */
.btn-checkin {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.quiz-limit-pill {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 7px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 9px;
  background: #f3f4f6;
  color: #6b7280;
  font-size: 12px;
  font-weight: 700;
}
/* item-mark: colored icon container instead of letter initial */
.item-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: rgba(94,182,157,0.12);
  color: #5eb69d;
  font-size: 14px;
  font-weight: 800;
}
.icon-button {
  border: 0;
  border-radius: 8px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 16px;
}
.icon-button.danger {
  background: #fff0ec;
  color: #b45242;
}
.submission-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}
.submission-row input {
  padding: 7px;
  font-size: 10px;
}
.learning-item {
  align-items: flex-start;
}
.item-main {
  min-width: 0;
}
.item-main h3 {
  margin-bottom: 5px;
}
.item-main p {
  line-height: 1.5;
}
@media (max-width: 650px) {
  .toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
  .form-grid {
    grid-template-columns: 1fr;
  }
  .form-grid .full {
    grid-column: auto;
  }
  .teacher-picker {
    grid-template-columns: 1fr;
  }
  .teacher-picker legend,
  .teacher-picker small {
    grid-column: auto;
  }
  .form-actions {
    flex-wrap: wrap;
  }
  .learning-item {
    flex-wrap: wrap;
  }
  .item-actions {
    width: 100%;
    margin-left: 55px;
  }
  .submission-row {
    align-items: flex-start;
    flex-direction: column;
  }
}
.detail-picker {
  display: flex;
  gap: 8px;
  margin: 12px 0;
}
.detail-picker select {
  flex: 1;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 11px 12px;
  background: white;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.detail-panel {
  margin: 18px 0;
}

/* =========================================================
   DETAIL MODAL (Floating Dialog)
   ========================================================= */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}
.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: scale(0.97) translateY(12px);
}

.detail-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(8px);
  overflow-y: auto;
}

.detail-modal-container {
  width: 100%;
  max-width: 680px;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  box-shadow: 0 24px 48px -12px rgba(15, 23, 42, 0.25), 0 0 0 1px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  animation: modalScaleIn 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalScaleIn {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Modal Header */
.detail-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 22px 24px 18px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  gap: 16px;
}

.detail-header-profile {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.detail-avatar {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 800;
  font-family: 'Manrope', sans-serif;
  color: #ffffff;
  flex-shrink: 0;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.1);
}
.detail-avatar.avatar-students {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}
.detail-avatar.avatar-teachers {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}
.detail-avatar.avatar-classes {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
}
.detail-avatar.avatar-subjects {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}
.detail-avatar.avatar-majors {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
}
.detail-avatar.avatar-materials {
  background: linear-gradient(135deg, #ec4899 0%, #db2777 100%);
}
.detail-avatar.avatar-assignments {
  background: linear-gradient(135deg, #f97316 0%, #ea580c 100%);
}
.detail-avatar.avatar-quizzes {
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
}

.detail-header-info {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-badge-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.detail-type-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #475569;
}
.detail-type-pill.pill-students {
  background: #d1fae5;
  color: #065f46;
}
.detail-type-pill.pill-teachers {
  background: #dbeafe;
  color: #1e40af;
}
.detail-type-pill.pill-classes {
  background: #ede9fe;
  color: #5b21b6;
}
.detail-type-pill.pill-subjects {
  background: #fef3c7;
  color: #92400e;
}
.detail-type-pill.pill-majors {
  background: #cffafe;
  color: #155e75;
}
.detail-type-pill.pill-materials {
  background: #fce7f3;
  color: #9d174d;
}
.detail-type-pill.pill-assignments {
  background: #ffedd5;
  color: #9a3412;
}

.detail-class-pill {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  border: 1px solid #bfdbfe;
}

.detail-unassigned-pill {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fecaca;
}

.detail-modal-title {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
  margin: 0;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-modal-subtitle {
  font-size: 12px;
  color: #64748b;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.email-sep {
  color: #cbd5e1;
  font-weight: 800;
}

/* Modal Body */
.detail-modal-body {
  padding: 22px 24px;
  overflow-y: auto;
  max-height: calc(88vh - 150px);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-card-group {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 16px 18px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 9px;
  margin-bottom: 14px;
  color: #1e293b;
}

.group-header h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  color: #1e293b;
  letter-spacing: -0.01em;
}

.group-icon {
  color: #3b82f6;
  flex-shrink: 0;
}

.info-grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 16px;
}

.info-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-cell.full-width {
  grid-column: 1 / -1;
}

.info-label {
  font-size: 11px;
  font-weight: 600;
  color: #64748b;
}

.info-value {
  font-size: 13px;
  font-weight: 600;
  color: #1e293b;
  word-break: break-word;
}

.info-value-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-copy {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: 1px solid #cbd5e1;
  border-radius: 7px;
  background: white;
  color: #475569;
  cursor: pointer;
  transition: all 0.15s ease;
  padding: 0;
  flex-shrink: 0;
}
.btn-copy:hover {
  background: #f1f5f9;
  border-color: #94a3b8;
  color: #0f172a;
}

.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.02em;
}

.inline-icon {
  display: inline-block;
  vertical-align: -2px;
  color: #94a3b8;
  margin-right: 4px;
}

.status-active-text {
  color: #059669;
  font-weight: 700;
}

/* Student Specific */
.student-class-highlight {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-left: 4px solid #3b82f6;
  border-radius: 10px;
  padding: 12px 16px;
}

.class-info-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.class-main-name {
  font-size: 14px;
  font-weight: 700;
  color: #0f172a;
}

.class-sub-meta {
  font-size: 11px;
  color: #64748b;
  font-weight: 500;
}

.btn-open-class {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 700;
  color: #2563eb;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  padding: 6px 12px;
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.btn-open-class:hover {
  background: #dbeafe;
  border-color: #93c5fd;
}

.student-no-class-alert {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: #fffbeb;
  border: 1px solid #fef3c7;
  border-radius: 10px;
  padding: 12px 14px;
  color: #92400e;
}
.student-no-class-alert strong {
  display: block;
  font-size: 12px;
  margin-bottom: 2px;
}
.student-no-class-alert p {
  margin: 0;
  font-size: 11px;
  color: #b45309;
}
.alert-icon {
  color: #d97706;
  flex-shrink: 0;
  margin-top: 2px;
}

/* Teacher assignments */
.assignment-cards-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.assignment-mini-card {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 10px 12px;
}

.assignment-icon-wrap {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #eff6ff;
  color: #2563eb;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.assignment-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.assignment-content strong {
  font-size: 12px;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.assignment-content span {
  font-size: 11px;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-assignment-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  padding: 14px;
  color: #64748b;
  font-size: 12px;
}

/* Chips */
.chip-grid-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.person-badge-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 5px 10px;
  font-size: 11px;
  font-weight: 600;
  color: #334155;
}

.empty-hint-text {
  margin: 0;
  font-size: 12px;
  color: #94a3b8;
  font-style: italic;
}

/* Files attachment */
.files-attachment-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.file-download-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  color: #2563eb;
  cursor: pointer;
  transition: all 0.15s ease;
}
.file-download-chip:hover {
  background: #eff6ff;
  border-color: #93c5fd;
}

/* Assignment submission box */
.submission-box-highlight {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 14px 16px;
}

.submission-status-badge {
  align-self: flex-start;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  background: #ecfdf5;
  color: #059669;
}

.submission-time-text,
.submission-score-text {
  margin: 0;
  font-size: 12px;
  color: #475569;
}

.upload-action-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
  flex-wrap: wrap;
}

/* Modal Footer */
.detail-modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  gap: 16px;
}

.modal-footer-left {
  flex: 1;
}

.modal-footer-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.copied-toast {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  border-radius: 999px;
  color: #047857;
  font-size: 11px;
  font-weight: 700;
}

@media (max-width: 650px) {
  .detail-modal-header {
    padding: 18px 16px;
  }
  .detail-modal-body {
    padding: 16px;
  }
  .info-grid-2 {
    grid-template-columns: 1fr;
  }
  .assignment-cards-list {
    grid-template-columns: 1fr;
  }
  .student-class-highlight {
    flex-direction: column;
    align-items: flex-start;
  }
  .detail-modal-footer {
    flex-direction: column;
    align-items: stretch;
    padding: 14px 16px;
  }
  .modal-footer-right {
    justify-content: flex-end;
  }
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}
.detail-grid div {
  display: grid;
  gap: 5px;
  padding: 12px;
  border: 1px solid #e5ebe6;
  border-radius: 8px;
}
.detail-grid .detail-wide {
  grid-column: 1 / -1;
}
.detail-assignment-list > div {
  display: grid;
  gap: 6px;
}
.detail-assignment-list p {
  margin: 0;
  color: var(--ink);
  font-size: 11px;
  line-height: 1.5;
}
.detail-link-list {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  padding: 0 !important;
  border: 0 !important;
}
.detail-person-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--ink) !important;
  font-size: 11px !important;
}
.detail-link-action {
  color: var(--green);
  font-size: 10px;
  font-weight: 700;
  text-decoration: underline;
}
.detail-link {
  display: inline-flex;
  align-items: center;
  border: 1px solid #cfe1d5;
  border-radius: 7px;
  padding: 7px 10px;
  color: var(--green);
  background: #f3f8f4;
  font-size: 11px;
  font-weight: 700;
  text-decoration: none;
}
.detail-link span {
  color: var(--green);
  font-size: 10px;
  text-decoration: underline;
}
.detail-link:hover {
  border-color: var(--green);
  background: #e5f4e8;
}
.detail-grid span {
  color: var(--muted);
  font-size: 10px;
}
.detail-grid strong {
  font-size: 12px;
}
@media (max-width: 650px) {
  .detail-picker {
    align-items: stretch;
    flex-direction: column;
  }
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
.admin-card-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 18px;
}
.admin-data-card {
  padding: 18px;
  border: 1px solid #e1e9e3;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(26, 58, 39, 0.05);
}
.card-heading {
  display: flex;
  align-items: center;
  gap: 12px;
}
.card-heading h3 {
  margin: 0;
}
.admin-data-card > p {
  min-height: 36px;
  color: var(--muted);
  font-size: 11px;
  line-height: 1.5;
}
.card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 14px;
}
.button-danger {
  border: 0;
  border-radius: 8px;
  padding: 10px 14px;
  background: #fff0ec;
  color: #b45242;
  cursor: pointer;
  font: inherit;
  font-size: 11px;
  font-weight: 700;
}
.admin-card-grid ~ .item-list {
  display: none;
}
.material-groups {
  display: grid;
  gap: 16px;
  margin-top: 18px;
}
.material-group {
  padding: 18px;
}
.material-group-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.material-group-heading h3 {
  margin: 0;
}
.teacher-badge {
  padding: 7px 10px;
  border-radius: 999px;
  background: #e5f4e8;
  color: #34704b;
  font-size: 10px;
  font-weight: 700;
}
.material-group-items {
  display: grid;
  gap: 8px;
  margin-top: 14px;
}
.material-group-items .learning-item {
  border-top: 1px solid var(--line);
  padding-top: 13px;
}
.resource-context {
  display: block;
  margin-top: 8px;
  color: var(--muted);
  font-size: 10px;
}
.resource-files {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
}
.file-name,
.file-link,
.file-delete,
.resource-links a {
  color: var(--green);
  font-size: 10px;
}
.file-link,
.file-delete {
  border: 0;
  padding: 0;
  background: transparent;
  cursor: pointer;
  text-decoration: underline;
}
.file-delete {
  color: #b45242;
}
.resource-links {
  display: flex;
  gap: 10px;
  margin-top: 9px;
}
.submission-recap {
  margin-top: 18px;
}
.submission-table {
  overflow-x: auto;
  margin-top: 14px;
}
.submission-table-row {
  display: grid;
  grid-template-columns: 0.35fr 0.8fr 1.4fr 1.25fr 1.4fr 0.7fr;
  min-width: 850px;
  gap: 10px;
  padding: 10px 0;
  border-top: 1px solid var(--line);
  font-size: 11px;
}
.submission-table-row:not(.submission-table-head) {
  cursor: pointer;
}
.submission-table-row:not(.submission-table-head):hover {
  background: #f3f8f4;
}
.submission-detail-panel {
  margin-top: 18px;
}
.submission-detail-panel label {
  display: grid;
  gap: 7px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 600;
}
.submission-detail-panel input,
.submission-detail-panel textarea {
  width: 100%;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 9px 10px;
  background: #fff;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.submission-table-head {
  color: var(--muted);
  font-size: 10px;
  font-weight: 700;
}
.report-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  padding-bottom: 16px;
  border-bottom: 2px solid #2f6048;
}
.report-overline {
  margin: 0 0 5px;
  color: #2f6048;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.08em;
}
.report-heading h2 {
  margin: 0 0 8px;
  color: #1f3427;
  font-size: 20px;
}
.report-meta {
  margin: 3px 0;
  color: var(--muted);
  font-size: 11px;
}
@media (max-width: 650px) {
  .report-heading {
    align-items: stretch;
    flex-direction: column;
  }
}
@media print {
  :global(body.printing-report *) {
    visibility: hidden !important;
  }
  :global(body.printing-report .submission-recap),
  :global(body.printing-report .submission-recap *) {
    visibility: visible !important;
  }
  :global(body.printing-report .submission-recap) {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    margin: 0;
    padding: 18px;
    border: 0;
    box-shadow: none;
  }
  :global(body.printing-report .recap-actions),
  :global(body.printing-report .submission-detail-panel) {
    display: none !important;
  }
  .report-heading {
    border-bottom-color: #222;
  }
  .submission-table-row {
    min-width: 0;
    grid-template-columns: 0.35fr 0.8fr 1.4fr 1.25fr 1.4fr 0.7fr;
    font-size: 10px;
  }
}
@media print {
  .submission-recap {
    break-inside: avoid;
  }
  .submission-recap ~ * {
    display: none !important;
  }
}
@media (max-width: 700px) {
  .admin-card-grid {
    grid-template-columns: 1fr;
  }
}

.attendance-stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}
.stat-mini-card {
  background: white;
  border: 1px solid #edf0f2;
  border-radius: 18px;
  padding: 18px 16px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(15,23,42,0.04);
  transition: transform 0.2s;
}
.stat-mini-card:hover { transform: translateY(-2px); }
.stat-mini-icon {
  margin-bottom: 8px;
  opacity: 0.6;
}
.stat-mini-val {
  display: block;
  font-size: 24px;
  font-weight: 800;
  color: #111827;
  line-height: 1.1;
  margin-bottom: 4px;
  font-family: 'Poppins', sans-serif;
}
.stat-mini-lbl {
  font-size: 10px;
  font-weight: 700;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.stat-mini-card.stat-rate { border-top: 3px solid #5eb69d; }
.stat-mini-card.stat-rate .stat-mini-val { color: #5eb69d; }
.stat-mini-card.stat-rate .stat-mini-icon { color: #5eb69d; opacity: 1; }
.stat-mini-card.stat-present { border-top: 3px solid #10b981; }
.stat-mini-card.stat-present .stat-mini-val { color: #10b981; }
.stat-mini-card.stat-present .stat-mini-icon { color: #10b981; opacity: 1; }
.stat-mini-card.stat-excused { border-top: 3px solid #3b82f6; }
.stat-mini-card.stat-excused .stat-mini-val { color: #3b82f6; }
.stat-mini-card.stat-excused .stat-mini-icon { color: #3b82f6; opacity: 1; }
.stat-mini-card.stat-absent { border-top: 3px solid #f43f5e; }
.stat-mini-card.stat-absent .stat-mini-val { color: #f43f5e; }
.stat-mini-card.stat-absent .stat-mini-icon { color: #f43f5e; opacity: 1; }

.attendance-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.01em;
}
.attendance-pill.present {
  background: #ecfdf5;
  color: #059669;
  border: 1px solid #a7f3d0;
}
.attendance-pill.closed {
  background: #f3f4f6;
  color: #6b7280;
  border: 1px solid #e5e7eb;
}
/* Modernized learning item card */
.item-list .learning-item {
  background: white;
  border: 1px solid #e8ede9;
  border-radius: 18px;
  padding: 18px 20px;
  margin-bottom: 12px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.04);
  transition: box-shadow 0.2s, transform 0.2s;
  display: flex;
  align-items: flex-start;
  gap: 16px;
}
.item-list .learning-item .item-main {
  flex: 1 1 auto;
}
.item-list .learning-item .item-main > p {
  max-width: 72ch;
}
.item-list .learning-item:hover {
  box-shadow: 0 6px 20px rgba(0,0,0,0.08);
  transform: translateY(-1px);
}
.item-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin-top: 8px;
}
/* admin card enhanced */
.admin-data-card {
  padding: 20px;
  border: 1px solid #e1e9e3;
  border-radius: 18px;
  background: white;
  box-shadow: 0 4px 14px rgba(26,58,39,0.05);
  transition: box-shadow 0.2s, transform 0.2s;
}
.admin-data-card:hover {
  box-shadow: 0 8px 24px rgba(26,58,39,0.1);
  transform: translateY(-2px);
}

/* ================= DARK MODE STYLES ================= */
:global(html[data-theme='dark']) .tab-button {
  color: #94a3b8;
}
:global(html[data-theme='dark']) .tab-button:hover {
  background: rgba(101, 201, 171, 0.12);
  color: #f1f5f9;
}
:global(html[data-theme='dark']) .tab-button.active {
  background: #1e293b !important;
  color: var(--tab-color, #65c9ab) !important;
  border-color: var(--tab-color, #65c9ab) !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3) !important;
}

:global(html[data-theme='dark']) .item-list .learning-item,
:global(html[data-theme='dark']) .learning-item {
  background: #1e293b !important;
  border-color: #334155 !important;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25) !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .item-list .learning-item:hover,
:global(html[data-theme='dark']) .learning-item:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4) !important;
}
:global(html[data-theme='dark']) .item-main h3 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .item-main > p {
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .item-main .eyebrow {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .resource-context {
  color: #94a3b8 !important;
}

:global(html[data-theme='dark']) .resource-open,
:global(html[data-theme='dark']) .button-secondary {
  background: #273549 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .resource-open:hover,
:global(html[data-theme='dark']) .button-secondary:hover {
  background: #334155 !important;
}
:global(html[data-theme='dark']) .file-link,
:global(html[data-theme='dark']) .resource-links a {
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .button-danger {
  background: #3b1e22 !important;
  color: #fca5a5 !important;
}

:global(html[data-theme='dark']) .admin-data-card {
  background: #1e293b !important;
  border-color: #334155 !important;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25) !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .admin-data-card h3 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .admin-data-card > p {
  color: #94a3b8 !important;
}

:global(html[data-theme='dark']) .material-group {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .material-group-heading h3 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .teacher-badge {
  background: #164e3f !important;
  color: #a7f3d0 !important;
}
:global(html[data-theme='dark']) .material-group-items .learning-item {
  border-top-color: #334155 !important;
}

:global(html[data-theme='dark']) .stat-mini-card {
  background: #1e293b !important;
  border-color: #334155 !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2) !important;
}
:global(html[data-theme='dark']) .stat-mini-val {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .stat-mini-lbl {
  color: #94a3b8 !important;
}

:global(html[data-theme='dark']) .detail-grid div {
  background: #151f30 !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .detail-grid span {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .detail-grid strong {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .detail-assignment-list p {
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .teacher-picker {
  background: #151f30 !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .teacher-option {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .btn-icon-action {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .btn-icon-action:hover {
  background: #273549 !important;
}
:global(html[data-theme='dark']) .quiz-limit-pill {
  background: #273549 !important;
  border-color: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .detail-picker input,
:global(html[data-theme='dark']) .detail-picker select,
:global(html[data-theme='dark']) .form-grid input,
:global(html[data-theme='dark']) .form-grid textarea,
:global(html[data-theme='dark']) .form-grid select,
:global(html[data-theme='dark']) .submission-detail-panel input,
:global(html[data-theme='dark']) .submission-detail-panel textarea {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .submission-table-row {
  border-color: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .submission-table-row:not(.submission-table-head):hover {
  background: #273549 !important;
}
:global(html[data-theme='dark']) .submission-table-head {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .report-heading {
  border-bottom-color: #334155 !important;
}
:global(html[data-theme='dark']) .report-heading h2 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .report-overline {
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .report-meta {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .attendance-pill.closed {
  background: #273549 !important;
  border-color: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .empty-state,
:global(html[data-theme='dark']) .loading-state {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .banner-success {
  background: #064e3b !important;
  border-color: #059669 !important;
  color: #a7f3d0 !important;
}
:global(html[data-theme='dark']) .form-error {
  background: #451a1a !important;
  border-color: #7f1d1d !important;
  color: #fca5a5 !important;
}

/* Detail Modal Dark Mode */
:global(html[data-theme='dark']) .detail-modal-backdrop {
  background: rgba(15, 23, 42, 0.75);
}
:global(html[data-theme='dark']) .detail-modal-container {
  background: #0f172a !important;
  border-color: #334155 !important;
  box-shadow: 0 24px 48px -12px rgba(0, 0, 0, 0.6) !important;
}
:global(html[data-theme='dark']) .detail-modal-header {
  background: #1e293b !important;
  border-bottom-color: #334155 !important;
}
:global(html[data-theme='dark']) .detail-modal-title {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .detail-modal-subtitle {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .detail-card-group {
  background: #1e293b !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .group-header h4 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .info-label {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .info-value {
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .btn-copy {
  background: #0f172a !important;
  border-color: #475569 !important;
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .btn-copy:hover {
  background: #1e293b !important;
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .student-class-highlight {
  background: #0f172a !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .class-main-name {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .class-sub-meta {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .btn-open-class {
  background: #1e3a5f !important;
  border-color: #2563eb !important;
  color: #93c5fd !important;
}
:global(html[data-theme='dark']) .student-no-class-alert {
  background: #2a1f10 !important;
  border-color: #78350f !important;
  color: #fde68a !important;
}
:global(html[data-theme='dark']) .student-no-class-alert p {
  color: #fcd34d !important;
}
:global(html[data-theme='dark']) .assignment-mini-card {
  background: #0f172a !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .assignment-content strong {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .assignment-content span {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .assignment-icon-wrap {
  background: #1e3a5f !important;
  color: #93c5fd !important;
}
:global(html[data-theme='dark']) .person-badge-chip {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .file-download-chip {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #93c5fd !important;
}
:global(html[data-theme='dark']) .submission-box-highlight {
  background: #0f172a !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .submission-time-text,
:global(html[data-theme='dark']) .submission-score-text {
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .detail-modal-footer {
  background: #1e293b !important;
  border-top-color: #334155 !important;
}
:global(html[data-theme='dark']) .detail-type-pill {
  background: #334155 !important;
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .detail-class-pill {
  background: #1e3a5f !important;
  border-color: #2563eb !important;
  color: #93c5fd !important;
}
:global(html[data-theme='dark']) .detail-unassigned-pill {
  background: #3b1111 !important;
  border-color: #7f1d1d !important;
  color: #fca5a5 !important;
}
:global(html[data-theme='dark']) .empty-assignment-box {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}
</style>

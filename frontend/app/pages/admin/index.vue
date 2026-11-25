<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useApi } from "~/composables/useApi";
import { useAuthStore } from "~/stores/auth";

definePageMeta({ middleware: "auth" });
const auth = useAuthStore();
const { request } = useApi();
const route = useRoute();

type ClassDetail = { students: any[]; subjects: any[] };

const classes = ref<any[]>([]);
const classDetails = ref<Record<string, ClassDetail>>({});
const subjects = ref<any[]>([]);
const teachers = ref<any[]>([]);
const selectedClassId = ref("");
const studentId = ref("");
const selectedStudent = ref<any | null>(null);
const subjectId = ref("");
const selectedSubject = ref<any | null>(null);
const selectedTeacherId = ref("");
const studentMatches = ref<any[]>([]);
const subjectMatches = ref<any[]>([]);
const loading = ref(true);
const detailLoading = ref(false);
const errorMessage = ref("");
const actionMessage = ref("");

const idOf = (item: any) => item?.id || item?.ID;
const nameOf = (item: any) =>
  item?.name ||
  item?.Name ||
  item?.full_name ||
  item?.FullName ||
  item?.username ||
  item?.Username ||
  "Tanpa nama";

// Digunakan khusus untuk objek SISWA (punya field Classes/classes), bukan objek kelas.
const studentClasses = (student: any) =>
  student?.Classes ||
  student?.classes ||
  student?.Students ||
  student?.students ||
  [];
const studentClassNames = (student: any) =>
  studentClasses(student)
    .map((item: any) => nameOf(item))
    .join(", ");
const studentHasOtherClass = (student: any) =>
  studentClasses(student).some(
    (item: any) => idOf(item) !== selectedClassId.value,
  );

const selectedClass = computed(() =>
  classes.value.find((item) => idOf(item) === selectedClassId.value),
);
const members = computed(
  () => classDetails.value[selectedClassId.value]?.students || [],
);
const classSubjects = computed(
  () => classDetails.value[selectedClassId.value]?.subjects || [],
);

// --- Data ringkasan per kelas (dipakai di kartu roster) ---
const classStudents = (classItem: any) =>
  classDetails.value[idOf(classItem)]?.students || [];
const classSubjectsFor = (classItem: any) =>
  classDetails.value[idOf(classItem)]?.subjects || [];
const namesPreview = (
  items: any[],
  nameFn: (item: any) => string,
  limit = 5,
) => {
  if (!items.length) return "";
  const shown = items.slice(0, limit).map(nameFn);
  const rest = items.length - shown.length;
  return rest > 0 ? `${shown.join(", ")}, +${rest} lainnya` : shown.join(", ");
};
const studentNamesPreview = (classItem: any) =>
  namesPreview(classStudents(classItem), nameOf);
const subjectNamesPreview = (classItem: any) =>
  namesPreview(classSubjectsFor(classItem), (item) =>
    nameOf(item.Subject || item.subject || item),
  );

const fetchClassDetail = async (classId: string) => {
  if (!classId) return;
  const [memberResponse, subjectResponse] = await Promise.all([
    request<{ data?: any[] }>(`/admin/classes/${classId}/students`),
    request<{ data?: any[] }>(`/admin/classes/${classId}/subjects/detail`),
  ]);
  classDetails.value = {
    ...classDetails.value,
    [classId]: {
      students: memberResponse.data || [],
      subjects: subjectResponse.data || [],
    },
  };
};

const load = async () => {
  loading.value = true;
  errorMessage.value = "";
  try {
    const [classResponse, subjectResponse] = await Promise.all([
      request<{ data?: any[] }>("/admin/classes"),
      request<{ data?: any[] }>("/admin/subjects"),
    ]);
    classes.value = classResponse.data || [];
    subjects.value = subjectResponse.data || [];

    const userResponse = await request<{ data?: any[] }>("/admin/users");
    teachers.value = (userResponse.data || []).filter(
      (item: any) =>
        (item.role?.name || item.Role?.Name || "").toLowerCase() === "guru",
    );

    const requestedClassId = String(route.query.class_id || "");
    selectedClassId.value = classes.value.some(
      (item) => idOf(item) === requestedClassId,
    )
      ? requestedClassId
      : selectedClassId.value || idOf(classes.value[0]) || "";

    // Ambil detail (siswa + mapel) untuk SEMUA kelas sekaligus, agar kartu roster lengkap.
    await Promise.all(
      classes.value.map((item) => fetchClassDetail(idOf(item))),
    );
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Data administrasi belum dapat dimuat.";
  } finally {
    loading.value = false;
  }
};

const selectClass = (classId: string) => {
  selectedClassId.value = classId;
};

watch(selectedClassId, async (id) => {
  if (id && !classDetails.value[id]) {
    detailLoading.value = true;
    try {
      await fetchClassDetail(id);
    } finally {
      detailLoading.value = false;
    }
  }
});

// --- Pencarian dengan debounce agar tidak memanggil API di setiap ketukan ---
let studentSearchTimer: ReturnType<typeof setTimeout> | undefined;
const searchStudents = () => {
  selectedStudent.value = null;
  const query = studentId.value.trim();
  if (studentSearchTimer) clearTimeout(studentSearchTimer);
  if (query.length < 2) {
    studentMatches.value = [];
    return;
  }
  studentSearchTimer = setTimeout(async () => {
    try {
      const response = await request<{ data?: any[] }>(
        `/admin/students/search?q=${encodeURIComponent(query)}`,
      );
      studentMatches.value = response.data || [];
    } catch {
      studentMatches.value = [];
    }
  }, 300);
};

let subjectSearchTimer: ReturnType<typeof setTimeout> | undefined;
const searchSubjects = () => {
  selectedSubject.value = null;
  selectedTeacherId.value = "";
  const query = subjectId.value.trim();
  if (subjectSearchTimer) clearTimeout(subjectSearchTimer);
  if (query.length < 2) {
    subjectMatches.value = [];
    return;
  }
  subjectSearchTimer = setTimeout(async () => {
    try {
      const response = await request<{ data?: any[] }>(
        `/admin/subjects/search?q=${encodeURIComponent(query)}`,
      );
      subjectMatches.value = response.data || [];
    } catch {
      subjectMatches.value = [];
    }
  }, 300);
};

const chooseStudent = (student: any) => {
  selectedStudent.value = student;
  studentId.value = student.nis || student.NIS || student.id || student.ID;
  studentMatches.value = [];
};
const chooseSubject = (subject: any) => {
  selectedSubject.value = subject;
  selectedTeacherId.value = "";
  subjectId.value = subject.code || subject.Code || subject.id || subject.ID;
  subjectMatches.value = [];
};

type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
const execute = async (path: string, method: HttpMethod, body?: any) => {
  errorMessage.value = "";
  actionMessage.value = "";
  try {
    await request(path, { method, body });
    actionMessage.value = "Perubahan berhasil disimpan.";
    await fetchClassDetail(selectedClassId.value);
    return true;
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Perubahan tidak dapat disimpan.";
    return false;
  }
};

const addMember = async () => {
  const id = studentId.value.trim();
  if (!id) return;
  const ok = await execute(
    `/admin/classes/${selectedClassId.value}/students`,
    "POST",
    {
      student_id: id,
    },
  );
  // Input hanya dikosongkan jika berhasil, supaya data yang diketik tidak hilang saat gagal.
  if (ok) {
    studentId.value = "";
    selectedStudent.value = null;
    studentMatches.value = [];
  }
};
const removeMember = async (item: any) => {
  if (confirm("Keluarkan siswa dari kelas ini?")) {
    await execute(
      `/admin/classes/${selectedClassId.value}/students/${idOf(item)}`,
      "DELETE",
    );
  }
};
const addSubject = async () => {
  const id = subjectId.value.trim();
  if (!id || !selectedTeacherId.value) return;
  const ok = await execute(
    `/admin/classes/${selectedClassId.value}/subjects`,
    "POST",
    {
      subject_id: id,
      teacher_id: selectedTeacherId.value,
    },
  );
  if (ok) {
    subjectId.value = "";
    selectedSubject.value = null;
    selectedTeacherId.value = "";
    subjectMatches.value = [];
  }
};
const removeSubject = async (item: any) => {
  if (confirm("Hapus mata pelajaran dari kelas ini?")) {
    await execute(
      `/admin/classes/${selectedClassId.value}/subjects/${idOf(item)}`,
      "DELETE",
    );
  }
};
const assignTeacher = async (item: any, teacherId: string) => {
  if (!teacherId) return;
  await execute(
    `/admin/classes/${selectedClassId.value}/subjects/${idOf(item)}/teacher`,
    "PUT",
    { teacher_id: teacherId },
  );
};
const eligibleTeachers = (item: any) =>
  item.Subject?.Teachers ||
  item.subject?.Teachers ||
  item.Subject?.teachers ||
  item.subject?.teachers ||
  [];
const selectedSubjectTeachers = computed(
  () =>
    selectedSubject.value?.Teachers || selectedSubject.value?.teachers || [],
);

onMounted(load);
</script>

<template>
  <AppShell
    active="admin"
    eyebrow="School Space"
    title="Administrasi sekolah"
  >
    <section class="page-intro">
      <p class="eyebrow">Manajemen Pengguna & Kelas</p>
      <h2>Kelola struktur kelas dan anggotanya.</h2>
      <p>
        Klik salah satu kartu kelas di bawah, atau pilih dari dropdown, untuk
        melihat & mengelola detailnya.
      </p>
    </section>

    <div v-if="errorMessage" class="form-error">{{ errorMessage }}</div>
    <div v-if="actionMessage" class="success">{{ actionMessage }}</div>

    <div v-if="loading" class="loading-state">
      <span class="loader" /> Memuat administrasi...
    </div>
    <template v-else>
      <div class="class-picker">
        <label>
          Kelas aktif
          <select v-model="selectedClassId">
            <option
              v-for="item in classes"
              :key="idOf(item)"
              :value="idOf(item)"
            >
              {{ nameOf(item) }}
            </option>
          </select>
        </label>
        <NuxtLink class="button button-secondary" to="/learning"
          >Kelola pengguna & kelas</NuxtLink
        >
      </div>

      <section class="panel class-roster-panel">
        <div class="panel-heading">
          <div>
            <p class="eyebrow">Class roster</p>
            <h3>Ringkasan tiap kelas</h3>
          </div>
        </div>
        <div v-if="classes.length" class="class-roster-grid">
          <article
            v-for="classItem in classes"
            :key="`roster-${idOf(classItem)}`"
            class="class-roster-card"
            :class="{ active: idOf(classItem) === selectedClassId }"
            role="button"
            tabindex="0"
            @click="selectClass(idOf(classItem))"
            @keyup.enter="selectClass(idOf(classItem))"
          >
            <div class="roster-card-head">
              <strong>{{ nameOf(classItem) }}</strong>
              <div class="roster-badges">
                <span class="badge"
                  >{{ classStudents(classItem).length }} siswa</span
                >
                <span class="badge badge-alt"
                  >{{ classSubjectsFor(classItem).length }} mapel</span
                >
              </div>
            </div>
            <div class="roster-list">
              <p class="roster-list-title">Siswa terdaftar</p>
              <span v-if="classStudents(classItem).length">{{
                studentNamesPreview(classItem)
              }}</span>
              <small v-else>Belum ada siswa terdaftar.</small>
            </div>
            <div class="roster-list">
              <p class="roster-list-title">Mata pelajaran</p>
              <span v-if="classSubjectsFor(classItem).length">{{
                subjectNamesPreview(classItem)
              }}</span>
              <small v-else>Belum ada mata pelajaran.</small>
            </div>
          </article>
        </div>
        <div v-else class="empty-state">Belum ada kelas.</div>
      </section>

      <section class="admin-grid">
        <article class="panel">
          <div class="panel-heading">
            <div>
              <p class="eyebrow">Class members</p>
              <h3>
                {{ selectedClass ? nameOf(selectedClass) : "Pilih kelas" }}
              </h3>
            </div>
          </div>
          <div class="search-form">
            <input
              v-model="studentId"
              type="text"
              placeholder="Cari berdasarkan NIS atau nama siswa"
              @input="searchStudents"
              @keyup.enter="addMember"
            />
            <button
              class="button button-primary"
              type="button"
              :disabled="
                !studentId.trim() || studentHasOtherClass(selectedStudent)
              "
              @click="addMember"
            >
              Tambah
            </button>
          </div>
          <div v-if="studentMatches.length" class="search-preview">
            <button
              v-for="student in studentMatches"
              :key="idOf(student)"
              type="button"
              @click="chooseStudent(student)"
            >
              <strong>{{ student.full_name || student.FullName }}</strong>
              <span>
                NIS: {{ student.nis || student.NIS }}
                <template v-if="studentClassNames(student)">
                  · Kelas: {{ studentClassNames(student) }}</template
                >
              </span>
            </button>
          </div>
          <small class="field-hint"
            >Ketik minimal 2 karakter. Data harus cocok dengan database.</small
          >

          <div v-if="detailLoading" class="empty-state">
            Memuat data kelas...
          </div>
          <template v-else>
            <div
              v-for="member in members"
              :key="idOf(member)"
              class="compact-row"
            >
              <span>{{ nameOf(member) }}</span>
              <button
                class="icon-button danger"
                type="button"
                @click="removeMember(member)"
              >
                ×
              </button>
            </div>
            <div v-if="!members.length" class="empty-state">
              Belum ada anggota.
            </div>
          </template>
        </article>

        <article class="panel">
          <div class="panel-heading">
            <div>
              <p class="eyebrow">Class subjects</p>
              <h3>Mata pelajaran kelas</h3>
            </div>
          </div>
          <div class="search-form">
            <input
              v-model="subjectId"
              type="text"
              placeholder="Cari kode atau nama mata pelajaran"
              @input="searchSubjects"
              @keyup.enter="addSubject"
            />
            <button
              class="button button-primary"
              type="button"
              :disabled="!subjectId.trim() || !selectedTeacherId"
              @click="addSubject"
            >
              Tambah
            </button>
          </div>
          <div v-if="subjectMatches.length" class="search-preview">
            <button
              v-for="subject in subjectMatches"
              :key="idOf(subject)"
              type="button"
              @click="chooseSubject(subject)"
            >
              <strong>{{ subject.code || subject.Code }}</strong>
              <span>{{ subject.name || subject.Name }}</span>
            </button>
          </div>
          <label v-if="selectedSubject" class="teacher-select-label">
            Guru pengajar di kelas
            <select v-model="selectedTeacherId" required>
              <option value="">Pilih guru pengajar</option>
              <option
                v-for="teacher in selectedSubjectTeachers"
                :key="idOf(teacher)"
                :value="idOf(teacher)"
              >
                {{ nameOf(teacher) }}
              </option>
            </select>
            <small v-if="!selectedSubjectTeachers.length" class="field-hint"
              >Mapel ini belum memiliki guru pengajar.</small
            >
          </label>
          <small class="field-hint"
            >Ketik minimal 2 karakter. Data harus cocok dengan database.</small
          >

          <div v-if="detailLoading" class="empty-state">
            Memuat data kelas...
          </div>
          <template v-else>
            <div
              v-for="subject in classSubjects"
              :key="idOf(subject)"
              class="compact-row"
            >
              <span>{{
                nameOf(subject.Subject || subject.subject || subject)
              }}</span>
              <button
                class="icon-button danger"
                type="button"
                @click="
                  removeSubject(subject.Subject || subject.subject || subject)
                "
              >
                ×
              </button>
            </div>
            <div v-if="!classSubjects.length" class="empty-state">
              Belum ada mata pelajaran.
            </div>
          </template>
        </article>
      </section>
    </template>

    <section v-if="classSubjects.length" class="panel teacher-assignment-panel">
      <div class="panel-heading">
        <div>
          <p class="eyebrow">Teacher assignment</p>
          <h3>Guru per mata pelajaran kelas</h3>
        </div>
      </div>
      <div
        v-for="subject in classSubjects"
        :key="`assignment-${idOf(subject)}`"
        class="teacher-assignment-row"
      >
        <strong>{{
          nameOf(subject.Subject || subject.subject || subject)
        }}</strong>
        <div v-if="eligibleTeachers(subject).length" class="teacher-options">
          <label
            v-for="teacher in eligibleTeachers(subject)"
            :key="idOf(teacher)"
            class="teacher-option"
          >
            <input
              type="radio"
              :name="`teacher-${idOf(subject)}`"
              :checked="
                idOf(subject.Teacher || subject.teacher) === idOf(teacher)
              "
              @change="assignTeacher(subject, idOf(teacher))"
            />
            <span>{{ nameOf(teacher) }}</span>
          </label>
        </div>
        <small v-else class="field-hint"
          >Atur guru yang dapat mengajar mapel ini melalui CRUD Mata
          Pelajaran.</small
        >
      </div>
    </section>
  </AppShell>
</template>

<style scoped>
.class-picker {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}
.class-picker label {
  width: min(360px, 100%);
}
select,
.search-form input {
  width: 100%;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 11px 12px;
  background: white;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.admin-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}
.class-roster-panel {
  margin-bottom: 18px;
}
.class-roster-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 10px;
}
.class-roster-card {
  display: grid;
  gap: 9px;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 14px;
  background: #fff;
  font-size: 11px;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background 0.15s ease;
}
.class-roster-card:hover {
  border-color: #a9c9b3;
}
.class-roster-card.active {
  border-color: #75aa84;
  background: #f3f8f4;
}
.roster-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.roster-card-head strong {
  font-size: 13px;
}
.roster-badges {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
.badge {
  padding: 2px 8px;
  border-radius: 999px;
  background: #eef3ef;
  color: #4c6b57;
  font-size: 10px;
  white-space: nowrap;
}
.badge-alt {
  background: #eef1fb;
  color: #4c5a8f;
}
.roster-list-title {
  margin: 0 0 3px;
  color: var(--ink);
  font-weight: 600;
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.roster-list span {
  color: var(--muted);
  line-height: 1.55;
}
.roster-list small {
  color: var(--muted);
}
.search-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  gap: 8px;
  margin-bottom: 6px;
}
.search-form input:focus,
.teacher-select-label select:focus {
  outline: none;
  border-color: #75aa84;
  box-shadow: 0 0 0 3px rgba(117, 170, 132, 0.14);
}
.search-preview {
  display: grid;
  gap: 3px;
  margin-bottom: 8px;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  background: #fff;
  overflow: hidden;
}
.search-preview button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 0;
  border-bottom: 1px solid #edf1ee;
  padding: 10px 12px;
  background: #fff;
  color: var(--ink);
  cursor: pointer;
  text-align: left;
  font: inherit;
  font-size: 11px;
}
.search-preview button:last-child {
  border-bottom: 0;
}
.search-preview button:hover {
  background: #f3f8f4;
}
.search-preview span {
  color: var(--muted);
  font-size: 10px;
}
.field-hint {
  display: block;
  margin: 0 0 14px;
  color: var(--muted);
  font-size: 10px;
}
.compact-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid var(--line);
  padding: 13px 0;
  font-size: 12px;
}
.icon-button {
  border: 0;
  border-radius: 8px;
  padding: 6px 10px;
  background: #fff0ec;
  color: #b45242;
  cursor: pointer;
  font-size: 16px;
}
.success {
  margin-bottom: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  background: #e5f4e8;
  color: #34704b;
  font-size: 11px;
}
@media (max-width: 700px) {
  .class-picker {
    align-items: stretch;
    flex-direction: column;
  }
  .admin-grid {
    grid-template-columns: 1fr;
  }
  .search-form {
    grid-template-columns: 1fr;
  }
}
.teacher-assignment-panel {
  margin-top: 18px;
}
.teacher-assignment-row {
  display: grid;
  gap: 8px;
  border-top: 1px solid var(--line);
  padding: 13px 0;
  font-size: 12px;
}
.teacher-options {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.teacher-option {
  display: flex !important;
  grid-template-columns: none !important;
  align-items: center;
  gap: 6px;
  color: var(--ink) !important;
  font-size: 11px !important;
  font-weight: 500 !important;
}
.teacher-option input {
  width: auto;
}
.teacher-assignment-row .field-hint {
  margin: 0;
}

/* ================= DARK MODE STYLES ================= */
:global(html[data-theme='dark']) select,
:global(html[data-theme='dark']) .search-form input {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .class-roster-card {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .class-roster-card:hover {
  background: #273549 !important;
  border-color: #65c9ab !important;
}
:global(html[data-theme='dark']) .class-roster-card.active {
  background: #1e3a32 !important;
  border-color: #65c9ab !important;
}
:global(html[data-theme='dark']) .teacher-assignment-row {
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .teacher-option {
  color: #f1f5f9 !important;
}
</style>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useApi } from "~/composables/useApi";
import { useAuthStore } from "~/stores/auth";

definePageMeta({ middleware: "auth" });

const auth = useAuthStore();
const { request } = useApi();
const active = ref("");
const years = ref<any[]>([]);
const classes = ref<any[]>([]);
const agendas = ref<any[]>([]);
const loading = ref(true);
const saving = ref(false);
const semesterSaving = ref(false);
const showForm = ref(false);
const editingId = ref("");
const expandedYearId = ref("");
const editingSemesterId = ref("");
const errorMessage = ref("");
const message = ref("");
const form = ref({ name: "", is_active: false, start_date: "", end_date: "" });
const semesterForm = ref({ number: 1, start_date: "", end_date: "" });
const agendaForm = ref({
  title: "",
  description: "",
  start_date: "",
  end_date: "",
  category: "Kegiatan sekolah",
  audience: "Semua",
  location: "",
  start_time: "",
  end_time: "",
  status: "Terjadwal",
});
const agendaYearId = ref("");
const editingAgendaId = ref("");
const agendaSaving = ref(false);
const calendarCursor = ref(new Date());
const selectedDate = ref("");
const selectedAgenda = ref<any | null>(null);

// This page owns the school's academic *timeline* — years and their semesters.
// Class/subject/major records themselves are managed on the school administration
// page; here they only ever appear read-only, grouped by the year they belong to.
const tabs = computed(() =>
  auth.role === "admin"
    ? [
        { key: "years", label: "Tahun akademik" },
        { key: "class-distribution", label: "Distribusi kelas" },
        { key: "timeline", label: "Timeline" },
      ]
    : [
        { key: "years", label: "Tahun akademik" },
        { key: "timeline", label: "Timeline & Agenda" },
      ],
);

const idOf = (item: any) => item.id || item.ID;
const labelOf = (item: any) =>
  item.name || item.Name || item.title || item.Title || "Tanpa nama";
const semestersOf = (year: any) => year.semesters || year.Semesters || [];
const agendasForYear = (year: any) =>
  agendas.value.filter((item) => {
    if ((item.academic_year_id || item.AcademicYearID) !== idOf(year))
      return false;
    if (auth.role === "admin") return true;
    const audience = String(
      item.audience || item.Audience || "Semua",
    ).toLowerCase();
    if (audience === "semua") return true;
    if (auth.role === "guru") return audience.includes("guru");
    if (auth.role === "siswa") return audience.includes("siswa");
    return true;
  });
const dateOnly = (value: string) => (value ? value.slice(0, 10) : "");
const agendaCategoryMeta = (category: string) => {
  const value = String(category || "").toLowerCase();
  if (value.includes("libur"))
    return { color: "red", icon: "●", label: "Libur" };
  if (value.includes("acara"))
    return { color: "green", icon: "◆", label: "Acara penting" };
  if (
    value.includes("akademik") ||
    value.includes("ujian") ||
    value.includes("pembelajaran")
  )
    return { color: "blue", icon: "■", label: "Agenda akademik" };
  if (value.includes("deadline") || value.includes("pengingat"))
    return { color: "yellow", icon: "▲", label: "Deadline / pengingat" };
  if (value.includes("sekolah") || value.includes("rapat"))
    return { color: "purple", icon: "★", label: "Kegiatan sekolah" };
  return { color: "gray", icon: "●", label: "Agenda lainnya" };
};
const calendarMonthLabel = computed(() =>
  calendarCursor.value.toLocaleDateString("id-ID", {
    month: "long",
    year: "numeric",
  }),
);
const calendarDays = computed(() => {
  const cursor = calendarCursor.value;
  const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1);
  const start = new Date(first);
  start.setDate(first.getDate() - ((first.getDay() + 6) % 7));
  return Array.from({ length: 42 }, (_, index) => {
    const date = new Date(start);
    date.setDate(start.getDate() + index);
    const iso = dateInput(date);
    return {
      iso,
      day: date.getDate(),
      inMonth: date.getMonth() === cursor.getMonth(),
      today: iso === today,
      agendas: agendas.value.filter(
        (agenda: any) =>
          iso >= dateOnly(agenda.start_date || agenda.StartDate) &&
          iso <= dateOnly(agenda.end_date || agenda.EndDate),
      ),
    };
  });
});
const selectedDateAgendas = computed(() =>
  agendas.value.filter((agenda: any) => {
    const date = selectedDate.value;
    return (
      date &&
      date >= dateOnly(agenda.start_date || agenda.StartDate) &&
      date <= dateOnly(agenda.end_date || agenda.EndDate)
    );
  }),
);
const changeMonth = (offset: number) => {
  const next = new Date(calendarCursor.value);
  next.setMonth(next.getMonth() + offset);
  calendarCursor.value = next;
};
const selectCalendarDate = (date: any) => {
  selectedDate.value = date.iso;
  selectedAgenda.value = date.agendas[0] || null;
};
const agendaTime = (agenda: any) => {
  const start = agenda.start_time || agenda.StartTime;
  const end = agenda.end_time || agenda.EndTime;
  return start ? `${start}${end ? ` - ${end}` : ""}` : "Sepanjang hari";
};
const openAgendaForDate = () => {
  const year = years.value.find(
    (item: any) =>
      selectedDate.value >= yearStart(item) &&
      selectedDate.value <= yearEnd(item),
  );
  if (!year || auth.role !== "admin" || isArchivedYear(year)) return;
  resetAgendaForm(year);
  agendaForm.value.start_date = selectedDate.value;
  agendaForm.value.end_date = selectedDate.value;
};
const selectAgenda = (agenda: any) => {
  selectedAgenda.value = agenda;
  selectedDate.value = dateOnly(agenda.start_date || agenda.StartDate);
};
const isActiveYear = (year: any) => Boolean(year.is_active ?? year.IsActive);
const isArchivedYear = (year: any) =>
  Boolean(year.is_archived ?? year.IsArchived);
const yearStart = (year?: any) =>
  dateOnly(year?.start_date || year?.StartDate);
const yearEnd = (year?: any) => dateOnly(year?.end_date || year?.EndDate);
const semesterRange = (year: any) => {
  const semesters = semestersOf(year)
    .filter((semester: any) =>
      dateOnly(semester.start_date || semester.StartDate) &&
      dateOnly(semester.end_date || semester.EndDate),
    )
    .sort(
      (left: any, right: any) =>
        dateOnly(left.start_date || left.StartDate).localeCompare(
          dateOnly(right.start_date || right.StartDate),
        ),
    );

  return {
    start: semesters.length
      ? dateOnly(semesters[0].start_date || semesters[0].StartDate)
      : yearStart(year),
    end: semesters.length
      ? dateOnly(
          semesters[semesters.length - 1].end_date ||
            semesters[semesters.length - 1].EndDate,
        )
      : yearEnd(year),
  };
};
const academicYearPosition = (year: any) => {
  const orderedYears = [...years.value].sort((left: any, right: any) => {
    const startComparison = yearStart(left).localeCompare(yearStart(right));
    return startComparison || labelOf(left).localeCompare(labelOf(right));
  });
  const position = orderedYears.findIndex(
    (item: any) => idOf(item) === idOf(year),
  );
  return position >= 0 ? position + 1 : "-";
};
const activeSelectionWarning = computed(() => {
  const current = activeYear.value;
  if (
    !form.value.is_active ||
    !current ||
    current.id === editingId.value ||
    current.ID === editingId.value
  )
    return "";
  return `Tahun ${labelOf(current)} saat ini aktif dan akan dinonaktifkan.`;
});

const activeYear = computed(() => years.value.find(isActiveYear) || null);
const today = new Date().toISOString().slice(0, 10);
const activeSemester = computed(() => {
  return (
    years.value
      .flatMap((year) => semestersOf(year))
      .find((semester: any) => semester.status === "ongoing") || null
  );
});
const daysUntilSemesterEnd = computed(() => {
  if (!activeSemester.value) return null;
  const end = dateOnly(
    activeSemester.value.end_date || activeSemester.value.EndDate,
  );
  const remaining = Math.ceil(
    (new Date(`${end}T23:59:59`).getTime() - Date.now()) / 86400000,
  );
  return remaining >= 0 && remaining <= 14 ? remaining : null;
});
const semesterReminder = computed(() =>
  daysUntilSemesterEnd.value === null
    ? ""
    : daysUntilSemesterEnd.value === 0
      ? "Semester berakhir hari ini."
      : `Semester berakhir dalam ${daysUntilSemesterEnd.value} hari.`,
);
// Only years with a start/end range AND a span greater than zero days are safe to
// plot on the timeline — a same-day range would divide by zero when placing bars.
const timelineYears = computed(() =>
  years.value.filter((year) => {
    const start = yearStart(year);
    const end = yearEnd(year);
    if (!start || !end) return false;
    return new Date(end).getTime() > new Date(start).getTime();
  }),
);

const semesterBarStyle = (year: any, semester: any) => {
  const yearStartMs = new Date(yearStart(year)).getTime();
  const yearEndMs = new Date(yearEnd(year)).getTime();
  const span = yearEndMs - yearStartMs;
  const semStartMs = new Date(
    dateOnly(semester.start_date || semester.StartDate),
  ).getTime();
  const semEndMs = new Date(
    dateOnly(semester.end_date || semester.EndDate),
  ).getTime();
  const left = Math.min(
    97,
    Math.max(0, ((semStartMs - yearStartMs) / span) * 100),
  );
  const width = Math.min(
    100 - left,
    Math.max(3, ((semEndMs - semStartMs) / span) * 100),
  );
  return { left: `${left}%`, width: `${width}%` };
};

// Groups classes by their academic year so an admin can see, at a glance, which
// classes belong to which year — and surface any class that was never linked to
// a year at all, a gap the plain class list on the other page can't reveal.
const classesByYearId = computed(() => {
  const map = new Map<string, any[]>();
  for (const classItem of classes.value) {
    const yearId =
      classItem.academic_year_id ||
      classItem.AcademicYearID ||
      classItem.academic_year?.id ||
      classItem.AcademicYear?.ID ||
      "";
    const key = yearId || "unassigned";
    if (!map.has(key)) map.set(key, []);
    map.get(key)!.push(classItem);
  }
  return map;
});
const classesForYear = (year: any) =>
  classesByYearId.value.get(idOf(year)) || [];
const unassignedClasses = computed(
  () => classesByYearId.value.get("unassigned") || [],
);
const classGrade = (classItem: any) =>
  classItem.grade || classItem.Grade || "-";
const classMajor = (classItem: any) =>
  classItem.major?.name || classItem.Major?.Name || "Tanpa jurusan";

const resetForm = (year?: any) => {
  editingId.value = year ? idOf(year) : "";
  form.value = {
    name: year?.name || year?.Name || "",
    is_active: year?.is_active ?? year?.IsActive ?? false,
    start_date: yearStart(year),
    end_date: yearEnd(year),
  };
  showForm.value = true;
};

const resetSemesterForm = (semester?: any) => {
  editingSemesterId.value = semester ? idOf(semester) : "";
  semesterForm.value = {
    number: semester?.number || semester?.Number || 1,
    start_date: dateOnly(semester?.start_date || semester?.StartDate),
    end_date: dateOnly(semester?.end_date || semester?.EndDate),
  };
};

const resetAgendaForm = (year: any, agenda?: any) => {
  agendaYearId.value = idOf(year);
  editingAgendaId.value = agenda ? idOf(agenda) : "";
  agendaForm.value = {
    title: agenda?.title || agenda?.Title || "",
    description: agenda?.description || agenda?.Description || "",
    start_date:
      dateOnly(agenda?.start_date || agenda?.StartDate) || yearStart(year),
    end_date: dateOnly(agenda?.end_date || agenda?.EndDate) || yearStart(year),
    category: agenda?.category || agenda?.Category || "Kegiatan sekolah",
    audience: agenda?.audience || agenda?.Audience || "Semua",
    location: agenda?.location || agenda?.Location || "",
    start_time: agenda?.start_time || agenda?.StartTime || "",
    end_time: agenda?.end_time || agenda?.EndTime || "",
    status: agenda?.status || agenda?.Status || "Terjadwal",
  };
};
const closeAgendaForm = () => {
  editingAgendaId.value = "";
  agendaYearId.value = "";
};

const dateInput = (date: Date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
};
const prepareSemesterForm = (year: any) => {
  const existing = semestersOf(year);
  const number = existing.length
    ? Math.max(
        ...existing.map((item: any) => Number(item.number || item.Number || 0)),
      ) + 1
    : 1;
  const start = new Date(`${yearStart(year)}T00:00:00`);
  const end = new Date(`${yearEnd(year)}T00:00:00`);
  const semesterStart = new Date(start);
  semesterStart.setMonth(start.getMonth() + (number > 1 ? 6 : 0));
  const semesterEnd = new Date(start);
  semesterEnd.setMonth(start.getMonth() + (number > 1 ? 11 : 5));
  return {
    number,
    start_date: dateInput(semesterStart < end ? semesterStart : end),
    end_date: dateInput(semesterEnd < end ? semesterEnd : end),
  };
};

const semesterModalYear = ref<any | null>(null);
const showSemesterModal = ref(false);
const collapsedYears = ref<Record<string, boolean>>({});

const isYearCollapsed = (year: any) => Boolean(collapsedYears.value[idOf(year)]);
const toggleYearCollapse = (year: any) => {
  const yId = idOf(year);
  collapsedYears.value[yId] = !collapsedYears.value[yId];
};

const formatIndoDate = (val?: string) => {
  if (!val) return "-";
  const str = val.slice(0, 10);
  const parts = str.split("-");
  if (parts.length !== 3) return val;
  const d = new Date(Number(parts[0]), Number(parts[1]) - 1, Number(parts[2]));
  return d.toLocaleDateString("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
};

const semesterTitle = (semester: any) => {
  const num = Number(semester?.number || semester?.Number || 1);
  if (num === 1) return "Semester 1 (Ganjil)";
  if (num === 2) return "Semester 2 (Genap)";
  return `Semester ${num}`;
};

const semesterDuration = (semester: any) => {
  const startStr = dateOnly(semester?.start_date || semester?.StartDate);
  const endStr = dateOnly(semester?.end_date || semester?.EndDate);
  if (!startStr || !endStr) return { weeks: 0, days: 0, text: "-" };
  const start = new Date(`${startStr}T00:00:00`);
  const end = new Date(`${endStr}T23:59:59`);
  const diffTime = Math.max(0, end.getTime() - start.getTime());
  const days = Math.round(diffTime / (1000 * 60 * 60 * 24));
  const weeks = Math.round(days / 7);
  return { weeks, days, text: `${weeks} Minggu (${days} Hari)` };
};

const semesterProgress = (semester: any) => {
  const startStr = dateOnly(semester?.start_date || semester?.StartDate);
  const endStr = dateOnly(semester?.end_date || semester?.EndDate);
  if (!startStr || !endStr) return 0;
  const start = new Date(`${startStr}T00:00:00`).getTime();
  const end = new Date(`${endStr}T23:59:59`).getTime();
  const now = Date.now();
  if (now <= start) return 0;
  if (now >= end) return 100;
  return Math.min(
    100,
    Math.max(0, Math.round(((now - start) / (end - start)) * 100)),
  );
};

const semesterCountdown = (semester: any) => {
  const status = semester?.status || "unknown";
  const startStr = dateOnly(semester?.start_date || semester?.StartDate);
  const endStr = dateOnly(semester?.end_date || semester?.EndDate);
  const now = Date.now();

  if (status === "ongoing") {
    if (!endStr) return "";
    const end = new Date(`${endStr}T23:59:59`).getTime();
    const remainingDays = Math.ceil((end - now) / (1000 * 60 * 60 * 24));
    if (remainingDays <= 0) return "Berakhir hari ini";
    return `Sisa ${remainingDays} hari lagi`;
  }
  if (status === "coming") {
    if (!startStr) return "";
    const start = new Date(`${startStr}T00:00:00`).getTime();
    const daysUntil = Math.ceil((start - now) / (1000 * 60 * 60 * 24));
    if (daysUntil <= 0) return "Dimulai segera";
    return `Dimulai dalam ${daysUntil} hari`;
  }
  if (status === "ended") {
    return "Telah berakhir";
  }
  return "";
};

const agendasForSemester = (year: any, semester: any) => {
  const yearAgendas = agendasForYear(year);
  const semStart = dateOnly(semester?.start_date || semester?.StartDate);
  const semEnd = dateOnly(semester?.end_date || semester?.EndDate);
  if (!semStart || !semEnd) return [];
  return yearAgendas.filter((agenda: any) => {
    const aStart = dateOnly(agenda?.start_date || agenda?.StartDate);
    const aEnd = dateOnly(agenda?.end_date || agenda?.EndDate);
    return aStart <= semEnd && aEnd >= semStart;
  });
};

const openEditSemester = (year: any, semester: any) => {
  errorMessage.value = "";
  semesterModalYear.value = year;
  editingSemesterId.value = idOf(semester);
  semesterForm.value = {
    number: Number(semester?.number || semester?.Number || 1),
    start_date: dateOnly(semester?.start_date || semester?.StartDate),
    end_date: dateOnly(semester?.end_date || semester?.EndDate),
  };
  showSemesterModal.value = true;
};

const openAddSemester = (year: any) => {
  errorMessage.value = "";
  semesterModalYear.value = year;
  editingSemesterId.value = "";
  semesterForm.value = prepareSemesterForm(year);
  showSemesterModal.value = true;
};

const closeSemesterModal = () => {
  showSemesterModal.value = false;
  editingSemesterId.value = "";
  semesterModalYear.value = null;
  resetSemesterForm();
};

const handleSaveSemester = async () => {
  if (!semesterModalYear.value) return;
  await saveSemester(semesterModalYear.value);
};

const toggleYear = (year: any) => {
  const opening = expandedYearId.value !== idOf(year);
  expandedYearId.value = opening ? idOf(year) : "";
  if (opening && !editingSemesterId.value && !isArchivedYear(year))
    semesterForm.value = prepareSemesterForm(year);
};

const load = async () => {
  loading.value = true;
  errorMessage.value = "";
  try {
    let yearList: any[] = [];
    if (auth.role === "admin") {
      const yearResponse = await request<{ data?: any[] }>("/admin/academic-years");
      yearList = yearResponse.data || [];
    } else {
      try {
        const yearResponse = await request<{ data?: any[] }>("/academic-years");
        yearList = Array.isArray(yearResponse.data) ? yearResponse.data : [];
      } catch {
        const activeResponse = await request<{ data?: any }>("/academic-years/active").catch(() => null);
        yearList = activeResponse?.data ? [activeResponse.data] : [];
      }
    }
    const classResponse =
      auth.role === "admin"
        ? await request<{ data?: any[] }>("/admin/classes").catch(() => ({ data: [] }))
        : { data: [] };
    years.value = yearList;
    classes.value = classResponse.data || [];
    const agendaResponses = await Promise.all(
      years.value.map((year) =>
        request<{ data?: any[] }>(
          `/academic-years/${idOf(year)}/agendas`,
        ).catch(() => ({ data: [] })),
      ),
    );
    agendas.value = agendaResponses.flatMap((response) => response.data || []);
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Data akademik belum dapat dimuat.";
  } finally {
    loading.value = false;
  }
};

const saveAgenda = async () => {
  const year = years.value.find((item) => idOf(item) === agendaYearId.value);
  if (!year) return;
  errorMessage.value = "";
  if (!agendaForm.value.title.trim()) {
    errorMessage.value = "Judul agenda wajib diisi.";
    return;
  }
  if (
    !agendaForm.value.start_date ||
    !agendaForm.value.end_date ||
    agendaForm.value.end_date < agendaForm.value.start_date
  ) {
    errorMessage.value =
      "Tanggal selesai agenda harus sama atau setelah tanggal mulai.";
    return;
  }
  if (
    agendaForm.value.start_date < yearStart(year) ||
    agendaForm.value.end_date > yearEnd(year)
  ) {
    errorMessage.value =
      "Rentang agenda harus berada di dalam rentang tahun akademik.";
    return;
  }
  agendaSaving.value = true;
  try {
    const agendaId = editingAgendaId.value;
    const endpoint = agendaId
      ? `/admin/academic-years/${agendaYearId.value}/agendas/${agendaId}`
      : `/admin/academic-years/${agendaYearId.value}/agendas`;
    const body = {
      ...agendaForm.value,
      start_date: `${agendaForm.value.start_date}T00:00:00Z`,
      end_date: `${agendaForm.value.end_date}T23:59:59Z`,
    };
    await request(endpoint, { method: agendaId ? "PUT" : "POST", body });
    message.value = agendaId ? "Agenda diperbarui." : "Agenda berhasil dibuat.";
    closeAgendaForm();
    await load();
  } catch (error: any) {
    errorMessage.value = error?.data?.message || "Agenda belum dapat disimpan.";
  } finally {
    agendaSaving.value = false;
  }
};
const removeAgenda = async (agenda: any) => {
  if (!confirm(`Hapus agenda ${agenda.title || agenda.Title}?`)) return;
  try {
    await request(
      `/admin/academic-years/${agendaYearId.value || agenda.academic_year_id || agenda.AcademicYearID}/agendas/${idOf(agenda)}`,
      { method: "DELETE" },
    );
    message.value = "Agenda dihapus.";
    await load();
  } catch (error: any) {
    errorMessage.value = error?.data?.message || "Agenda belum dapat dihapus.";
  }
};

const saveYear = async () => {
  saving.value = true;
  errorMessage.value = "";
  if (
    !form.value.start_date ||
    !form.value.end_date ||
    form.value.end_date <= form.value.start_date
  ) {
    errorMessage.value =
      "Tanggal selesai tahun akademik harus setelah tanggal mulai.";
    saving.value = false;
    return;
  }
  try {
    const endpoint = editingId.value
      ? `/admin/academic-years/${editingId.value}`
      : "/admin/academic-years";
    const body = {
      ...form.value,
      start_date: `${form.value.start_date}T00:00:00Z`,
      end_date: `${form.value.end_date}T23:59:59Z`,
    };
    await request(endpoint, { method: editingId.value ? "PUT" : "POST", body });
    message.value = editingId.value
      ? "Tahun akademik diperbarui."
      : "Tahun akademik dibuat.";
    showForm.value = false;
    await load();
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Tahun akademik belum dapat disimpan.";
  } finally {
    saving.value = false;
  }
};

const saveSemester = async (year: any) => {
  errorMessage.value = "";
  if (
    semesterForm.value.start_date &&
    semesterForm.value.end_date &&
    semesterForm.value.end_date <= semesterForm.value.start_date
  ) {
    errorMessage.value =
      "Tanggal selesai semester harus setelah tanggal mulai.";
    return;
  }
  if (
    semesterForm.value.start_date < yearStart(year) ||
    semesterForm.value.end_date > yearEnd(year)
  ) {
    errorMessage.value =
      "Rentang semester harus berada di dalam rentang tahun akademik.";
    return;
  }
  const editingId = editingSemesterId.value;
  const overlaps = semestersOf(year).some((semester: any) => {
    if (idOf(semester) === editingId) return false;
    const start = dateOnly(semester.start_date || semester.StartDate);
    const end = dateOnly(semester.end_date || semester.EndDate);
    return (
      start <= semesterForm.value.end_date &&
      end >= semesterForm.value.start_date
    );
  });
  if (overlaps) {
    errorMessage.value =
      "Rentang tanggal semester bertumpang tindih dengan semester lain.";
    return;
  }
  semesterSaving.value = true;
  try {
    const yearId = idOf(year);
    const semesterId = editingSemesterId.value;
    const endpoint = semesterId
      ? `/admin/academic-years/${yearId}/semesters/${semesterId}`
      : `/admin/academic-years/${yearId}/semesters`;
    const body = {
      ...semesterForm.value,
      start_date: `${semesterForm.value.start_date}T00:00:00Z`,
      end_date: `${semesterForm.value.end_date}T23:59:59Z`,
    };
    await request(endpoint, { method: semesterId ? "PUT" : "POST", body });
    message.value = semesterId
      ? "Semester diperbarui."
      : "Semester ditambahkan.";
    closeSemesterModal();
    await load();
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Semester belum dapat disimpan.";
  } finally {
    semesterSaving.value = false;
  }
};

const removeYear = async (year: any) => {
  const linkedClasses = classesForYear(year);
  if (linkedClasses.length) {
    errorMessage.value = `Tahun ${labelOf(year)} tidak dapat dihapus karena masih memiliki ${linkedClasses.length} kelas terhubung. Pindahkan kelas terlebih dahulu di Administrasi Sekolah.`;
    return;
  }
  if (!confirm(`Hapus ${labelOf(year)}?`)) return;
  try {
    await request(`/admin/academic-years/${idOf(year)}`, { method: "DELETE" });
    message.value = "Tahun akademik dihapus.";
    await load();
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message ||
      "Tahun akademik masih digunakan atau belum dapat dihapus.";
  }
};

const removeSemester = async (year: any, semester: any) => {
  if (!confirm(`Hapus Semester ${semester.number || semester.Number}?`)) return;
  try {
    await request(
      `/admin/academic-years/${idOf(year)}/semesters/${idOf(semester)}`,
      { method: "DELETE" },
    );
    message.value = "Semester dihapus.";
    await load();
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Semester belum dapat dihapus.";
  }
};

const duplicateYear = async (year: any) => {
  const nextName = prompt(
    "Nama tahun akademik baru:",
    `${labelOf(year)} salinan`,
  );
  if (!nextName?.trim()) return;
  try {
    await request(`/admin/academic-years/${idOf(year)}/duplicate`, {
      method: "POST",
      body: { name: nextName.trim() },
    });
    message.value = `Struktur ${labelOf(year)} berhasil diduplikasi.`;
    await load();
  } catch (error: any) {
    errorMessage.value =
      error?.data?.message || "Tahun akademik belum dapat diduplikasi.";
  }
};
const printCalendar = () => window.print();

onMounted(async () => {
  active.value = tabs.value[0]?.key || "years";
  await load();
});
</script>

<template>
  <AppShell
    active="academic"
    eyebrow="School Space"
    title="Kalender Akademik"
  >
    <section class="page-intro">
      <p class="eyebrow">{{ auth.role === 'admin' ? 'Tahun ajaran & semester' : 'Kalender Akademik' }}</p>
      <h2>{{ auth.role === 'admin' ? 'Atur garis waktu akademik sekolah.' : 'Garis waktu & agenda akademik sekolah.' }}</h2>
      <p v-if="auth.role === 'admin'">
        Halaman ini khusus untuk tahun ajaran dan semester — kapan tiap tahun
        ajaran berjalan, dan kelas mana saja yang terhubung ke tahun tersebut.
        Untuk menambah atau mengubah data kelas, jurusan, guru, siswa, dan mata
        pelajaran satu per satu, buka halaman
        <strong>Administrasi Sekolah</strong>.
      </p>
      <p v-else>
        Lihat daftar tahun akademik, periode semester aktif, dan timeline agenda kegiatan sekolah secara lengkap.
      </p>
    </section>

    <transition name="fade-slide">
      <div v-if="errorMessage" class="banner banner-error" role="alert">
        <span class="banner-icon">!</span>{{ errorMessage }}
        <button
          class="banner-dismiss"
          type="button"
          aria-label="Tutup"
          @click="errorMessage = ''"
        >
          ×
        </button>
      </div>
    </transition>
    <transition name="fade-slide">
      <div v-if="message" class="banner banner-success" role="status">
        <span class="banner-icon">✓</span>{{ message }}
        <button
          class="banner-dismiss"
          type="button"
          aria-label="Tutup"
          @click="message = ''"
        >
          ×
        </button>
      </div>
    </transition>

    <div class="toolbar">
      <div class="tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-button"
          role="tab"
          :aria-selected="active === tab.key"
          :class="{ active: active === tab.key }"
          type="button"
          @click="active = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="toolbar-actions">
        <AppButton
          v-if="active === 'years' && auth.role === 'admin'"
          @click="resetForm()"
          >+ Tahun akademik</AppButton
        >
        <AppButton
          v-if="active === 'timeline'"
          variant="secondary"
          @click="printCalendar"
          >Cetak / PDF</AppButton
        >
      </div>
    </div>

    <transition name="fade-slide">
      <section v-if="showForm && active === 'years' && auth.role === 'admin'" class="panel form-panel">
        <div class="panel-heading">
          <div>
            <p class="eyebrow">Academic year form</p>
            <h3>
              {{ editingId ? "Edit tahun akademik" : "Tambah tahun akademik" }}
            </h3>
          </div>
          <button
            class="close-button"
            type="button"
            aria-label="Tutup form"
            @click="showForm = false"
          >
            ×
          </button>
        </div>
        <form class="form-grid" @submit.prevent="saveYear">
          <label
            >Nama<input v-model="form.name" required placeholder="2026/2027"
          /></label>
          <label
            >Mulai<input v-model="form.start_date" type="date" required
          /></label>
          <label
            >Selesai<input v-model="form.end_date" type="date" required
          /></label>
          <label class="checkbox full"
            ><input v-model="form.is_active" type="checkbox" /> Jadikan tahun
            aktif</label
          >
          <p v-if="activeSelectionWarning" class="warning-text">
            ⚠ {{ activeSelectionWarning }}
          </p>
          <div class="form-actions">
            <AppButton
              variant="secondary"
              type="button"
              @click="showForm = false"
              >Batal</AppButton
            >
            <AppButton type="submit" :disabled="saving">{{
              saving ? "Menyimpan..." : "Simpan"
            }}</AppButton>
          </div>
        </form>
      </section>
    </transition>

    <!-- Modal Dialog Tambah / Edit Semester -->
    <transition name="fade-slide">
      <div
        v-if="showSemesterModal && semesterModalYear"
        class="modal-backdrop"
        @click.self="closeSemesterModal"
      >
        <div class="modal-card">
          <div class="modal-head">
            <div>
              <p class="eyebrow">Manajemen Semester</p>
              <h3>
                {{ editingSemesterId ? "Edit Data Semester" : "Tambah Semester Baru" }}
              </h3>
              <p class="modal-sub">
                Tahun Ajaran: <strong>{{ labelOf(semesterModalYear) }}</strong>
                ({{ formatIndoDate(yearStart(semesterModalYear)) }} –
                {{ formatIndoDate(yearEnd(semesterModalYear)) }})
              </p>
            </div>
            <button
              type="button"
              class="close-button"
              aria-label="Tutup modal"
              @click="closeSemesterModal"
            >
              ×
            </button>
          </div>

          <form class="modal-body" @submit.prevent="handleSaveSemester">
            <div class="form-field">
              <label class="field-label">Pilihan Semester</label>
              <div class="semester-picker-row">
                <button
                  type="button"
                  class="picker-btn"
                  :class="{ active: semesterForm.number === 1 }"
                  @click="semesterForm.number = 1"
                >
                  Semester 1 (Ganjil)
                </button>
                <button
                  type="button"
                  class="picker-btn"
                  :class="{ active: semesterForm.number === 2 }"
                  @click="semesterForm.number = 2"
                >
                  Semester 2 (Genap)
                </button>
                <div v-if="semesterForm.number > 2" class="custom-num-wrap">
                  <span class="custom-label">Semester ke-</span>
                  <input
                    v-model.number="semesterForm.number"
                    type="number"
                    min="1"
                    class="custom-num-input"
                    required
                  />
                </div>
              </div>
            </div>

            <div class="form-grid-2">
              <div class="form-field">
                <label class="field-label">Tanggal Mulai Semester</label>
                <input
                  v-model="semesterForm.start_date"
                  type="date"
                  required
                  class="field-input"
                />
                <small class="field-hint">
                  Awal TA: {{ formatIndoDate(yearStart(semesterModalYear)) }}
                </small>
              </div>
              <div class="form-field">
                <label class="field-label">Tanggal Selesai Semester</label>
                <input
                  v-model="semesterForm.end_date"
                  type="date"
                  required
                  class="field-input"
                />
                <small class="field-hint">
                  Batas akhir TA: {{ formatIndoDate(yearEnd(semesterModalYear)) }}
                </small>
              </div>
            </div>

            <div v-if="errorMessage" class="modal-error-alert">
              ⚠️ {{ errorMessage }}
            </div>

            <div class="modal-actions">
              <AppButton
                variant="secondary"
                type="button"
                @click="closeSemesterModal"
              >
                Batal
              </AppButton>
              <AppButton type="submit" :disabled="semesterSaving">
                {{
                  semesterSaving
                    ? "Menyimpan..."
                    : editingSemesterId
                      ? "Simpan Perubahan"
                      : "+ Tambahkan Semester"
                }}
              </AppButton>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <div v-if="loading" class="loading-state">
      <span class="loader" /> Memuat struktur akademik...
    </div>

    <section v-else-if="active === 'years'" class="item-list">
      <article
        v-for="year in years"
        :key="idOf(year)"
        class="learning-item year-card-container"
        :class="{ 'year-card-active': isActiveYear(year) }"
      >
        <div class="year-card-header">
          <div class="year-header-left">
            <div class="item-mark academic-year-mark">
              TA ke-{{ academicYearPosition(year) }}
            </div>
            <div class="year-title-block">
              <div class="year-eyebrow-row">
                <span class="eyebrow-text">Tahun Ajaran ke-{{ academicYearPosition(year) }}</span>
                <span class="meta-dot">·</span>
                <span class="year-period-text">
                  {{ formatIndoDate(semesterRange(year).start) }} – {{ formatIndoDate(semesterRange(year).end) }}
                </span>
              </div>
              <h3 class="item-title">
                {{ labelOf(year) }}
                <span class="badge-row">
                  <span v-if="isActiveYear(year)" class="active-badge">Aktif</span>
                  <span v-if="isArchivedYear(year)" class="archive-badge">Diarsipkan</span>
                </span>
              </h3>
            </div>
          </div>

          <div class="year-header-stats">
            <div class="stat-bubble">
              <span class="stat-bubble-val">{{ semestersOf(year).length }}</span>
              <span class="stat-bubble-lbl">Semester</span>
            </div>
            <div class="stat-bubble">
              <span class="stat-bubble-val">{{ classesForYear(year).length }}</span>
              <span class="stat-bubble-lbl">Kelas</span>
            </div>
            <div class="stat-bubble">
              <span class="stat-bubble-val">{{ agendasForYear(year).length }}</span>
              <span class="stat-bubble-lbl">Agenda</span>
            </div>
          </div>

          <div class="item-actions year-actions">
            <button
              v-if="!isArchivedYear(year) && auth.role === 'admin'"
              type="button"
              class="btn-action-primary"
              @click="openAddSemester(year)"
            >
              + Semester
            </button>
            <button
              type="button"
              class="btn-action-toggle"
              :class="{ collapsed: isYearCollapsed(year) }"
              :title="isYearCollapsed(year) ? 'Tampilkan semester' : 'Sembunyikan semester'"
              @click="toggleYearCollapse(year)"
            >
              <span>{{ isYearCollapsed(year) ? "Buka Kartu" : "Tutup Kartu" }}</span>
              <svg
                class="chevron-icon"
                :class="{ 'rotate-180': !isYearCollapsed(year) }"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="6 9 12 15 18 9" />
              </svg>
            </button>
            <template v-if="auth.role === 'admin'">
              <button
                v-if="!isArchivedYear(year)"
                type="button"
                class="btn-action-subtle"
                @click="resetForm(year)"
              >
                Edit TA
              </button>
              <button
                type="button"
                class="btn-action-subtle"
                @click="duplicateYear(year)"
              >
                Duplikat
              </button>
              <button
                v-if="!isArchivedYear(year)"
                type="button"
                class="link-action-danger"
                @click="removeYear(year)"
              >
                Hapus
              </button>
            </template>
          </div>
        </div>

        <!-- Semester Cards Grid Section -->
        <transition name="fade-slide">
          <div v-show="!isYearCollapsed(year)" class="semester-section">
            <div class="semester-section-header">
              <div>
                <h4 class="semester-section-title">Kartu Semester</h4>
                <p class="semester-section-desc">
                  Informasi periode, waktu berjalan, dan agenda kegiatan di tiap semester
                </p>
              </div>
              <span class="semester-count-chip">
                {{ semestersOf(year).length }} Semester Terdaftar
              </span>
            </div>

            <div class="semester-cards-grid">
              <!-- Semester Card -->
              <div
                v-for="semester in semestersOf(year)"
                :key="idOf(semester)"
                class="semester-card"
                :class="[
                  `semester-card--${semester.status || 'unknown'}`,
                  { 'is-current': semester.status === 'ongoing' },
                ]"
              >
                <!-- Card Head -->
                <div class="semester-card-head">
                  <div class="semester-tag">
                    <span class="semester-tag-badge">
                      <svg
                        width="13"
                        height="13"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
                        <path
                          d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"
                        />
                      </svg>
                      Semester {{ semester.number || semester.Number }}
                    </span>
                    <span
                      class="semester-status-pill"
                      :class="`status-${semester.status || 'unknown'}`"
                    >
                      <span
                        v-if="semester.status === 'ongoing'"
                        class="pulse-indicator"
                      />
                      <span
                        v-else-if="semester.status === 'ended'"
                        class="check-indicator"
                        >✓</span
                      >
                      <span v-else class="coming-indicator">⏳</span>
                      {{
                        semester.status_label ||
                        (semester.status === "ongoing"
                          ? "Sedang Berjalan"
                          : semester.status === "coming"
                            ? "Mendatang"
                            : "Selesai")
                      }}
                    </span>
                  </div>

                  <!-- Actions for Admin -->
                  <div
                    v-if="!isArchivedYear(year) && auth.role === 'admin'"
                    class="semester-card-actions"
                  >
                    <button
                      type="button"
                      class="card-action-btn"
                      title="Edit semester"
                      @click="openEditSemester(year, semester)"
                    >
                      <svg
                        width="12"
                        height="12"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <path
                          d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"
                        />
                      </svg>
                      Edit
                    </button>
                    <button
                      type="button"
                      class="card-action-btn card-action-btn--danger"
                      title="Hapus semester"
                      @click="removeSemester(year, semester)"
                    >
                      <svg
                        width="12"
                        height="12"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <polyline points="3 6 5 6 21 6" />
                        <path
                          d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                        />
                      </svg>
                      Hapus
                    </button>
                  </div>
                </div>

                <div class="semester-card-main">
                  <h4 class="semester-card-name">{{ semesterTitle(semester) }}</h4>

                  <!-- Date range & duration box -->
                  <div class="semester-date-box">
                    <div class="date-row">
                      <svg
                        class="date-icon"
                        width="14"
                        height="14"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                        <line x1="16" y1="2" x2="16" y2="6" />
                        <line x1="8" y1="2" x2="8" y2="6" />
                        <line x1="3" y1="10" x2="21" y2="10" />
                      </svg>
                      <span class="date-text">
                        {{ formatIndoDate(semester.start_date || semester.StartDate) }} —
                        {{ formatIndoDate(semester.end_date || semester.EndDate) }}
                      </span>
                    </div>
                    <div class="duration-badge">
                      {{ semesterDuration(semester).text }}
                    </div>
                  </div>

                  <!-- Progress / Timeline Tracker -->
                  <div class="semester-progress-card">
                    <div class="progress-card-top">
                      <span class="progress-title">
                        {{
                          semester.status === "ongoing"
                            ? "Waktu Berjalan"
                            : semester.status === "coming"
                              ? "Jadwal Mulai"
                              : "Status Semester"
                        }}
                      </span>
                      <span
                        class="progress-countdown"
                        :class="`countdown-${semester.status}`"
                      >
                        {{ semesterCountdown(semester) }}
                      </span>
                    </div>
                    <div class="progress-bar-bg">
                      <div
                        class="progress-bar-fill"
                        :class="`fill-${semester.status}`"
                        :style="{ width: `${semesterProgress(semester)}%` }"
                      />
                    </div>
                    <div class="progress-card-footer">
                      <span>{{ semesterProgress(semester) }}% masa semester</span>
                      <span
                        v-if="semester.status === 'ongoing'"
                        class="active-glow-text"
                        >● Aktif sekarang</span
                      >
                    </div>
                  </div>

                  <!-- Agendas in this Semester -->
                  <div class="semester-agendas">
                    <div class="agendas-header">
                      <span class="agendas-title">
                        <svg
                          width="12"
                          height="12"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                        >
                          <polygon
                            points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"
                          />
                        </svg>
                        Agenda Semester ({{ agendasForSemester(year, semester).length }})
                      </span>
                    </div>
                    <div
                      v-if="agendasForSemester(year, semester).length"
                      class="agenda-items-list"
                    >
                      <div
                        v-for="ag in agendasForSemester(year, semester).slice(0, 2)"
                        :key="idOf(ag)"
                        class="agenda-mini-chip"
                      >
                        <span
                          class="agenda-mini-dot"
                          :class="`dot-${agendaCategoryMeta(ag.category || ag.Category).color}`"
                        />
                        <span class="agenda-mini-title">{{
                          ag.title || ag.Title
                        }}</span>
                        <span class="agenda-mini-date">{{
                          formatIndoDate(ag.start_date || ag.StartDate)
                        }}</span>
                      </div>
                      <div
                        v-if="agendasForSemester(year, semester).length > 2"
                        class="agenda-more-note"
                      >
                        +{{ agendasForSemester(year, semester).length - 2 }} kegiatan lainnya
                      </div>
                    </div>
                    <p v-else class="agendas-empty-note">
                      Belum ada agenda terdaftar di rentang semester ini.
                    </p>
                  </div>
                </div>
              </div>

              <!-- Add Semester Card Button (Admin Only) -->
              <button
                v-if="!isArchivedYear(year) && auth.role === 'admin'"
                type="button"
                class="add-semester-card"
                @click="openAddSemester(year)"
              >
                <div class="add-card-circle">+</div>
                <h5 class="add-card-title">
                  Tambah
                  {{
                    semestersOf(year).length === 0
                      ? "Semester 1 (Ganjil)"
                      : semestersOf(year).length === 1
                        ? "Semester 2 (Genap)"
                        : "Semester Baru"
                  }}
                </h5>
                <p class="add-card-sub">
                  Atur rentang tanggal semester untuk melengkapi jadwal akademik
                </p>
              </button>

              <!-- Empty state when no semesters and not admin -->
              <div
                v-if="!semestersOf(year).length && auth.role !== 'admin'"
                class="empty-semesters-box"
              >
                <p>Belum ada semester yang dikonfigurasi untuk tahun akademik ini.</p>
              </div>
            </div>
          </div>
        </transition>
      </article>

      <div v-if="!years.length" class="empty-state">
        <p>Belum ada tahun akademik.</p>
        <AppButton
          v-if="auth.role === 'admin'"
          variant="secondary"
          @click="resetForm()"
          >+ Tambah tahun akademik</AppButton
        >
      </div>
    </section>

    <section v-else-if="active === 'class-distribution'" class="item-list">
      <article
        v-for="year in years"
        :key="idOf(year)"
        class="learning-item year-item"
      >
        <div class="item-mark academic-year-mark">
          TA ke-{{ academicYearPosition(year) }}
        </div>
        <div class="item-main">
          <p class="eyebrow">{{ classesForYear(year).length }} kelas</p>
          <h3 class="item-title">
            {{ labelOf(year) }}
            <span class="badge-row">
              <span v-if="isActiveYear(year)" class="active-badge">Aktif</span>
              <span v-if="isArchivedYear(year)" class="archive-badge"
                >Diarsipkan</span
              >
            </span>
          </h3>
          <div v-if="classesForYear(year).length" class="class-chip-row">
            <span
              v-for="classItem in classesForYear(year)"
              :key="idOf(classItem)"
              class="class-chip"
              >{{ labelOf(classItem) }} · Tingkat {{ classGrade(classItem) }} ·
              {{ classMajor(classItem) }}</span
            >
          </div>
          <p v-else class="empty-inline">
            Belum ada kelas yang terhubung ke tahun ajaran ini.
          </p>
        </div>
      </article>
      <article
        v-if="unassignedClasses.length"
        class="learning-item year-item warning-item"
      >
        <div class="item-mark">!</div>
        <div class="item-main">
          <p class="eyebrow">{{ unassignedClasses.length }} kelas</p>
          <h3 class="item-title">Belum terhubung ke tahun ajaran</h3>
          <div class="class-chip-row">
            <span
              v-for="classItem in unassignedClasses"
              :key="idOf(classItem)"
              class="class-chip class-chip-warning"
              >{{ labelOf(classItem) }} · Tingkat
              {{ classGrade(classItem) }}</span
            >
          </div>
        </div>
      </article>
      <div
        v-if="!years.length && !unassignedClasses.length"
        class="empty-state"
      >
        <p>Belum ada data kelas atau tahun akademik.</p>
      </div>
    </section>

    <section v-else-if="active === 'timeline'" class="timeline-panel panel">
      <div class="panel-heading">
        <div>
          <p class="eyebrow">Kalender interaktif</p>
          <h3>Agenda akademik</h3>
        </div>
        <AppButton variant="secondary" @click="printCalendar"
          >Cetak / PDF</AppButton
        >
      </div>
      <div class="calendar-toolbar">
        <button
          type="button"
          class="calendar-nav"
          aria-label="Bulan sebelumnya"
          @click="changeMonth(-1)"
        >
          ‹
        </button>
        <h4>{{ calendarMonthLabel }}</h4>
        <button
          type="button"
          class="calendar-nav"
          aria-label="Bulan berikutnya"
          @click="changeMonth(1)"
        >
          ›
        </button>
        <button
          type="button"
          class="today-button"
          @click="
            calendarCursor = new Date();
            selectedDate = today;
            selectedAgenda = selectedDateAgendas[0] || null;
          "
        >
          Hari ini
        </button>
      </div>
      <div class="calendar-weekdays">
        <span
          v-for="day in ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min']"
          :key="day"
          >{{ day }}</span
        >
      </div>
      <div class="calendar-grid">
        <button
          v-for="day in calendarDays"
          :key="day.iso"
          type="button"
          class="calendar-day"
          :class="{
            muted: !day.inMonth,
            today: day.today,
            selected: selectedDate === day.iso,
          }"
          @click="selectCalendarDate(day)"
        >
          <span class="calendar-day-number">{{ day.day }}</span>
          <span v-if="day.agendas.length" class="agenda-count"
            >{{ day.agendas.length }} agenda</span
          >
          <span class="calendar-dots"
            ><i
              v-for="category in Array.from(
                new Set(
                  day.agendas.map(
                    (agenda: any) =>
                      agendaCategoryMeta(agenda.category || agenda.Category)
                        .color,
                  ),
                ),
              )"
              :key="String(category)"
              :class="`dot-${category}`"
          /></span>
          <span
            v-for="agenda in day.agendas.slice(0, 2)"
            :key="idOf(agenda)"
            class="calendar-agenda-label"
            @click.stop="selectAgenda(agenda)"
            >{{ agenda.title || agenda.Title }}</span
          >
        </button>
      </div>
      <div class="calendar-legend">
        <span
          v-for="category in [
            { color: 'red', label: 'Libur' },
            { color: 'green', label: 'Acara penting' },
            { color: 'blue', label: 'Akademik / pembelajaran' },
            { color: 'yellow', label: 'Deadline / pengingat' },
            { color: 'purple', label: 'Kegiatan sekolah' },
            { color: 'gray', label: 'Lainnya' },
          ]"
          :key="category.color"
          ><i :class="`dot-${category.color}`" />{{ category.label }}</span
        >
      </div>
      <aside v-if="selectedDate" class="agenda-detail-panel" aria-live="polite">
        <div class="panel-heading">
          <div>
            <p class="eyebrow">{{ selectedDate }}</p>
            <h3>
              {{
                selectedAgenda
                  ? selectedAgenda.title || selectedAgenda.Title
                  : "Tidak ada agenda"
              }}
            </h3>
          </div>
          <button
            type="button"
            class="close-button"
            aria-label="Tutup detail"
            @click="
              selectedDate = '';
              selectedAgenda = null;
            "
          >
            ×
          </button>
        </div>
        <template v-if="selectedAgenda">
          <div class="agenda-detail-meta">
            <span
              >{{
                agendaCategoryMeta(
                  selectedAgenda.category || selectedAgenda.Category,
                ).icon
              }}
              {{ selectedAgenda.category || selectedAgenda.Category }}</span
            ><span>{{ agendaTime(selectedAgenda) }}</span
            ><span>{{
              selectedAgenda.status || selectedAgenda.Status || "Terjadwal"
            }}</span>
          </div>
          <RichTextContent
            :content="selectedAgenda.description || selectedAgenda.Description || ''"
          />
          <small v-if="selectedAgenda.location || selectedAgenda.Location"
            >Lokasi:
            {{ selectedAgenda.location || selectedAgenda.Location }}</small
          >
          <small v-if="selectedAgenda.CreatedBy || selectedAgenda.created_by"
            >Dibuat oleh:
            {{
              labelOf(selectedAgenda.CreatedBy || selectedAgenda.created_by)
            }}</small
          >
          <div v-if="selectedDateAgendas.length > 1" class="agenda-detail-list">
            <button
              v-for="agenda in selectedDateAgendas"
              :key="idOf(agenda)"
              type="button"
              :class="{ active: selectedAgenda === agenda }"
              @click="selectAgenda(agenda)"
            >
              {{ agenda.title || agenda.Title }}
            </button>
          </div>
          <div v-if="auth.role === 'admin'" class="agenda-actions">
            <button
              type="button"
              @click="
                resetAgendaForm(
                  years.find(
                    (year: any) =>
                      idOf(year) ===
                      (selectedAgenda.academic_year_id ||
                        selectedAgenda.AcademicYearID),
                  ),
                  selectedAgenda,
                )
              "
            >
              Edit</button
            ><button
              type="button"
              class="link-action-danger"
              @click="removeAgenda(selectedAgenda)"
            >
              Hapus
            </button>
          </div>
        </template>
        <template v-else
          ><p>Belum ada agenda pada tanggal ini.</p>
          <AppButton
            v-if="auth.role === 'admin'"
            variant="secondary"
            @click="openAgendaForDate"
            >+ Tambah agenda pada tanggal ini</AppButton
          ></template
        >
      </aside>
      <div
        v-for="year in timelineYears"
        :key="idOf(year)"
        class="timeline-year"
      >
        <div class="timeline-label">
          <strong>{{ labelOf(year) }}</strong>
          <span>{{ yearStart(year) }} – {{ yearEnd(year) }}</span>
        </div>
        <div class="timeline-track">
          <div
            class="year-bar"
            :class="{
              archived: isArchivedYear(year),
              active: isActiveYear(year),
            }"
          >
            <span
              v-for="semester in semestersOf(year)"
              :key="idOf(semester)"
              class="semester-bar"
              :style="semesterBarStyle(year, semester)"
              :title="`Semester ${semester.number || semester.Number}: ${dateOnly(semester.start_date || semester.StartDate)} – ${dateOnly(semester.end_date || semester.EndDate)}`"
              >S{{ semester.number || semester.Number }}</span
            >
          </div>
          <div class="agenda-list">
            <div
              v-for="agenda in agendasForYear(year)"
              :key="idOf(agenda)"
              class="agenda-row"
            >
              <button
                type="button"
                class="agenda-row-button"
                @click="selectAgenda(agenda)"
              >
                <strong>{{ agenda.title || agenda.Title }}</strong
                ><span
                  >{{ dateOnly(agenda.start_date || agenda.StartDate)
                  }}<template
                    v-if="
                      dateOnly(agenda.end_date || agenda.EndDate) !==
                      dateOnly(agenda.start_date || agenda.StartDate)
                    "
                  >
                    –
                    {{ dateOnly(agenda.end_date || agenda.EndDate) }}</template
                  >
                  · {{ agenda.category || agenda.Category }} ·
                  {{ agenda.audience || agenda.Audience }}</span
                >
              </button>
              <div
                v-if="auth.role === 'admin' && !isArchivedYear(year)"
                class="agenda-actions"
              >
                <button type="button" @click="resetAgendaForm(year, agenda)">
                  Edit</button
                ><button
                  type="button"
                  class="link-action-danger"
                  @click="removeAgenda(agenda)"
                >
                  Hapus
                </button>
              </div>
            </div>
            <p v-if="!agendasForYear(year).length" class="empty-inline">
              Belum ada agenda kegiatan.
            </p>
            <AppButton
              v-if="auth.role === 'admin' && !isArchivedYear(year)"
              variant="secondary"
              @click="resetAgendaForm(year)"
              >+ Agenda kegiatan</AppButton
            >
          </div>
        </div>
      </div>
      <div v-if="!timelineYears.length" class="empty-state">
        <p>Belum ada rentang tahun akademik untuk ditampilkan.</p>
      </div>
      <form
        v-if="agendaYearId && auth.role === 'admin'"
        class="agenda-form"
        @submit.prevent="saveAgenda"
      >
        <div class="panel-heading">
          <div>
            <p class="eyebrow">Agenda timeline</p>
            <h3>{{ editingAgendaId ? "Edit agenda" : "Tambah agenda" }}</h3>
          </div>
          <button type="button" class="close-button" @click="closeAgendaForm">
            ×
          </button>
        </div>
        <div class="agenda-form-grid">
          <label
            >Judul<input
              v-model="agendaForm.title"
              required
              placeholder="Rapat orang tua"
          /></label>
          <label
            >Kategori<select v-model="agendaForm.category">
              <option>Kegiatan sekolah</option>
              <option>Ujian</option>
              <option>Libur</option>
              <option>Rapat</option>
              <option>Acara</option>
            </select></label
          >
          <label
            >Untuk<select v-model="agendaForm.audience">
              <option>Semua</option>
              <option>Guru</option>
              <option>Siswa</option>
              <option>Guru dan siswa</option>
            </select></label
          >
          <label
            >Lokasi<input
              v-model="agendaForm.location"
              placeholder="Aula sekolah"
          /></label>
          <label
            >Mulai
            <span class="date-time-fields">
              <input v-model="agendaForm.start_date" type="date" required />
              <input
                v-model="agendaForm.start_time"
                type="time"
                aria-label="Waktu mulai"
              />
            </span>
          </label>
          <label
            >Selesai
            <span class="date-time-fields">
              <input v-model="agendaForm.end_date" type="date" required />
              <input
                v-model="agendaForm.end_time"
                type="time"
                aria-label="Waktu selesai"
              />
            </span>
          </label>
          <label
            >Status<select v-model="agendaForm.status">
              <option>Terjadwal</option>
              <option>Berlangsung</option>
              <option>Selesai</option>
              <option>Dibatalkan</option>
            </select></label
          >
          <label class="agenda-description"
            >Catatan<RichTextEditor
              v-model="agendaForm.description"
              placeholder="Susun catatan dengan paragraf, daftar, atau penekanan teks..."
            />
          </label>
        </div>
        <div class="form-actions">
          <AppButton variant="secondary" type="button" @click="closeAgendaForm"
            >Batal</AppButton
          ><AppButton type="submit" :disabled="agendaSaving">{{
            agendaSaving ? "Menyimpan..." : "Simpan agenda"
          }}</AppButton>
        </div>
      </form>
    </section>

    <section v-else class="item-list">
      <article class="learning-item year-item overview-item">
        <div class="item-mark">
          {{ activeYear ? `TA ke-${academicYearPosition(activeYear)}` : "-" }}
        </div>
        <div class="item-main">
          <p class="eyebrow">Tahun ajaran berjalan</p>
          <h3>
            {{
              activeYear ? labelOf(activeYear) : "Belum ada tahun ajaran aktif"
            }}
          </h3>
          <p v-if="activeSemester" class="semester-line">
            Semester {{ activeSemester.number || activeSemester.Number }} ·
            {{
              dateOnly(activeSemester.start_date || activeSemester.StartDate)
            }}
            sampai
            {{ dateOnly(activeSemester.end_date || activeSemester.EndDate) }}
          </p>
          <p v-if="semesterReminder" class="reminder-badge">
            ⏳ {{ semesterReminder }}
          </p>
          <p v-else-if="activeYear" class="empty-inline">
            Belum ada semester yang sedang berjalan untuk tahun ini.
          </p>
          <div
            v-if="activeYear && agendasForYear(activeYear).length"
            class="overview-agenda"
          >
            <strong>Agenda tahun ini</strong>
            <span
              v-for="agenda in agendasForYear(activeYear)"
              :key="idOf(agenda)"
              >{{ dateOnly(agenda.start_date || agenda.StartDate) }} ·
              {{ agenda.title || agenda.Title }} ·
              {{ agenda.audience || agenda.Audience }}</span
            >
          </div>
        </div>
      </article>
    </section>
  </AppShell>
</template>

<style scoped>
/* Layout & toolbar */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 24px 0 18px;
  flex-wrap: wrap;
}
.tabs {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: #f1f5f2;
  border-radius: 10px;
  overflow-x: auto;
}
.tab-button {
  border: 1px solid transparent;
  border-radius: 7px;
  padding: 9px 14px;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}
.tab-button:hover {
  color: var(--ink);
}
.tab-button.active {
  background: white;
  color: var(--green);
  box-shadow: 0 1px 3px rgba(20, 40, 30, 0.08);
}
.toolbar-actions {
  display: flex;
  gap: 8px;
}

/* Panels & forms */
.form-panel {
  margin-bottom: 18px;
}
.panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}
.close-button {
  border: 0;
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  font-size: 20px;
  line-height: 1;
  padding: 4px;
  border-radius: 6px;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}
.close-button:hover {
  background: #f1f5f2;
  color: var(--ink);
}
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 15px;
}
.form-grid select,
.form-grid textarea {
  width: 100%;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 11px 12px;
  background: white;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.form-grid label {
  display: grid;
  gap: 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--muted);
}
.form-grid input,
.semester-form input {
  width: 100%;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 11px 12px;
  background: white;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}
.form-grid input:focus,
.semester-form input:focus {
  outline: none;
  border-color: var(--green);
  box-shadow: 0 0 0 3px rgba(76, 150, 112, 0.15);
}
.checkbox {
  display: flex !important;
  flex-direction: row !important;
  align-items: center;
  gap: 8px;
}
.checkbox input {
  width: auto;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  grid-column: 1/-1;
}

/* Item cards */
.learning-item {
  transition:
    box-shadow 0.15s ease,
    transform 0.15s ease;
}
.item-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.badge-row {
  display: inline-flex;
  gap: 6px;
}
.item-actions {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-wrap: wrap;
}
.academic-year-mark {
  width: auto;
  min-width: 62px;
  padding: 0 8px;
  white-space: nowrap;
  font-size: 11px;
}
.item-actions button,
.cancel-link,
.link-action {
  border: 0;
  background: none;
  color: var(--green);
  cursor: pointer;
  font-size: 11px;
  font-weight: 700;
  padding: 4px 2px;
  border-radius: 4px;
}
.item-actions button:hover,
.link-action:hover {
  text-decoration: underline;
}
.link-action-danger {
  color: #b45242 !important;
}
.active-badge {
  display: inline-flex;
  align-items: center;
  padding: 5px 10px;
  border-radius: 999px;
  background: #e5f4e8;
  color: #34704b;
  font-size: 10px;
  font-weight: 700;
}
.archive-badge {
  display: inline-flex;
  align-items: center;
  padding: 5px 10px;
  border-radius: 999px;
  background: #eef0f1;
  color: #667078;
  font-size: 10px;
  font-weight: 700;
}

/* Banners */
.banner {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 12px 40px 12px 14px;
  border-radius: 10px;
  font-size: 11px;
  line-height: 1.5;
}
.banner-error {
  background: #fdeeec;
  color: #a8402f;
}
.banner-success {
  background: #e5f4e8;
  color: #34704b;
}
.banner-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.08);
  font-size: 10px;
  font-weight: 800;
  flex-shrink: 0;
}
.banner-dismiss {
  position: absolute;
  top: 8px;
  right: 8px;
  border: 0;
  background: transparent;
  color: inherit;
  opacity: 0.6;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  padding: 4px 6px;
  border-radius: 6px;
}
.banner-dismiss:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.06);
}
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Year & Semester Card System */
.year-card-container {
  display: flex;
  flex-direction: column;
  gap: 0;
  border-radius: 14px;
  background: #ffffff;
  border: 1px solid #e5ebe6;
  box-shadow: 0 1px 4px rgba(20, 40, 30, 0.04);
  transition: all 0.2s ease;
  overflow: hidden;
}
.year-card-container:hover {
  border-color: #d1ded4;
  box-shadow: 0 4px 14px rgba(20, 40, 30, 0.07);
}
.year-card-active {
  border-color: #8bc99a;
  box-shadow: 0 0 0 1px #8bc99a, 0 4px 16px rgba(76, 150, 112, 0.08);
}

.year-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  padding: 16px 20px;
  width: 100%;
}
.year-header-left {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
  flex: 1 1 320px;
}
.year-title-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.year-eyebrow-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--muted);
  flex-wrap: wrap;
}
.meta-dot {
  opacity: 0.5;
}
.year-period-text {
  font-weight: 600;
  color: var(--ink);
}

.year-header-stats {
  display: flex;
  align-items: center;
  gap: 8px;
}
.stat-bubble {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 999px;
  background: #f1f5f2;
  font-size: 11px;
  color: var(--ink);
}
.stat-bubble-val {
  font-weight: 800;
  color: var(--green);
}
.stat-bubble-lbl {
  color: var(--muted);
  font-size: 10px;
  text-transform: lowercase;
}

.year-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.btn-action-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--green);
  color: #fff !important;
  border: 0;
  padding: 6px 12px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  transition: opacity 0.15s ease;
}
.btn-action-primary:hover {
  opacity: 0.9;
  text-decoration: none !important;
}
.btn-action-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: #f1f5f2;
  color: var(--ink) !important;
  border: 1px solid #dfe7e1;
  padding: 6px 11px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}
.btn-action-toggle:hover {
  background: #e5ece6;
  text-decoration: none !important;
}
.chevron-icon {
  transition: transform 0.2s ease;
}
.rotate-180 {
  transform: rotate(180deg);
}
.btn-action-subtle {
  background: transparent;
  color: var(--muted);
  border: 0;
  font-size: 11px;
  font-weight: 700;
  padding: 5px 6px;
  cursor: pointer;
}
.btn-action-subtle:hover {
  color: var(--ink);
  text-decoration: underline;
}

/* Semester section & grid */
.semester-section {
  width: 100%;
  padding: 18px 20px 22px;
  background: #fafbfa;
  border-top: 1px solid #edf1ee;
}
.semester-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}
.semester-section-title {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  color: var(--ink);
}
.semester-section-desc {
  margin: 2px 0 0;
  font-size: 11px;
  color: var(--muted);
}
.semester-count-chip {
  padding: 3px 9px;
  border-radius: 999px;
  background: #eef3ef;
  color: #34704b;
  font-size: 10px;
  font-weight: 700;
}

.semester-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(310px, 1fr));
  gap: 14px;
}

/* Individual Semester Card */
.semester-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #ffffff;
  border: 1px solid #e3eae4;
  border-radius: 12px;
  padding: 16px;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
  position: relative;
}
.semester-card:hover {
  border-color: #c9d8cc;
  box-shadow: 0 4px 14px rgba(20, 40, 30, 0.06);
  transform: translateY(-2px);
}
.semester-card.is-current {
  border-color: #8bc99a;
  box-shadow: 0 0 0 1px #8bc99a, 0 4px 16px rgba(76, 150, 112, 0.08);
  background: linear-gradient(180deg, #f7fcf8 0%, #ffffff 42%);
}

.semester-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.semester-tag {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.semester-tag-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 800;
  color: var(--ink);
}
.semester-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 800;
  line-height: 1;
}
.pulse-indicator {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #16a34a;
  box-shadow: 0 0 0 0 rgba(22, 163, 74, 0.7);
  animation: pulse-dot 1.8s infinite;
}
@keyframes pulse-dot {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(22, 163, 74, 0.7);
  }
  70% {
    transform: scale(1);
    box-shadow: 0 0 0 6px rgba(22, 163, 74, 0);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(22, 163, 74, 0);
  }
}
.check-indicator {
  font-size: 10px;
}
.coming-indicator {
  font-size: 9px;
}

.semester-card-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}
.card-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 7px;
  border-radius: 6px;
  border: 1px solid #dfe7e1;
  background: #fff;
  color: var(--ink) !important;
  font-size: 10px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}
.card-action-btn:hover {
  background: #f1f5f2;
  border-color: #cbd7cf;
  text-decoration: none !important;
}
.card-action-btn--danger {
  color: #b45242 !important;
}
.card-action-btn--danger:hover {
  background: #fff0ec;
  border-color: #f7c9c0;
}

.semester-card-main {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.semester-card-name {
  margin: 0;
  font-size: 14px;
  font-weight: 800;
  color: var(--ink);
}

/* Date box */
.semester-date-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 11px;
  border-radius: 8px;
  background: #f8fbf9;
  border: 1px solid #e7eee9;
  font-size: 11px;
}
.date-row {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.date-icon {
  color: var(--muted);
  flex-shrink: 0;
}
.date-text {
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.duration-badge {
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 700;
  color: var(--muted);
  background: #ffffff;
  padding: 2px 7px;
  border-radius: 6px;
  border: 1px solid #dfe7e1;
}

/* Progress card */
.semester-progress-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 9px 11px;
  border-radius: 8px;
  background: #f8fafc;
  border: 1px solid #edf2f7;
}
.progress-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 10px;
}
.progress-title {
  font-weight: 700;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.progress-countdown {
  font-weight: 700;
}
.countdown-ongoing {
  color: #15803d;
}
.countdown-coming {
  color: #b45309;
}
.countdown-ended {
  color: #64748b;
}

.progress-bar-bg {
  width: 100%;
  height: 6px;
  border-radius: 999px;
  background: #e2e8f0;
  overflow: hidden;
}
.progress-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.4s ease;
}
.fill-ongoing {
  background: linear-gradient(90deg, #4c9670, #22c55e);
}
.fill-coming {
  background: #eab308;
}
.fill-ended {
  background: #94a3b8;
}
.progress-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 10px;
  color: var(--muted);
}
.active-glow-text {
  color: #15803d;
  font-weight: 700;
}

/* Agendas highlight in semester */
.semester-agendas {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 8px;
  border-top: 1px dashed #e2e8f0;
}
.agendas-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 10px;
  font-weight: 700;
  color: var(--muted);
}
.agendas-title {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.agenda-items-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.agenda-mini-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 6px;
  background: #ffffff;
  border: 1px solid #edf2f7;
  font-size: 10px;
}
.agenda-mini-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.agenda-mini-title {
  flex: 1;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.agenda-mini-date {
  color: var(--muted);
  flex-shrink: 0;
  font-size: 9px;
}
.agenda-more-note {
  font-size: 9px;
  color: var(--muted);
  font-weight: 600;
  padding-left: 2px;
}
.agendas-empty-note {
  margin: 0;
  font-size: 10px;
  color: var(--muted);
  font-style: italic;
}

/* Add Semester Card */
.add-semester-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 8px;
  border: 2px dashed #cbd5e1;
  border-radius: 12px;
  background: #f8fafc;
  padding: 24px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  min-height: 220px;
}
.add-semester-card:hover {
  border-color: var(--green);
  background: #f0fdf4;
  transform: translateY(-2px);
}
.add-card-circle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  font-size: 20px;
  font-weight: 700;
  color: var(--green);
  transition: all 0.2s ease;
}
.add-semester-card:hover .add-card-circle {
  background: var(--green);
  color: #fff;
  border-color: var(--green);
}
.add-card-title {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  color: var(--ink);
}
.add-card-sub {
  margin: 0;
  font-size: 11px;
  color: var(--muted);
  max-width: 220px;
  line-height: 1.4;
}

.empty-semesters-box {
  grid-column: 1 / -1;
  padding: 24px;
  text-align: center;
  color: var(--muted);
  font-size: 12px;
  border: 1px dashed #cbd5e1;
  border-radius: 10px;
  background: #fff;
}

/* Modal Dialog */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 1000;
}
.modal-card {
  width: 100%;
  max-width: 520px;
  background: #ffffff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  animation: modal-pop 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
@keyframes modal-pop {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(6px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
.modal-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 20px 24px;
  border-bottom: 1px solid #edf2f7;
}
.modal-head h3 {
  margin: 2px 0 4px;
  font-size: 17px;
  font-weight: 800;
  color: var(--ink);
}
.modal-sub {
  margin: 0;
  font-size: 11px;
  color: var(--muted);
}
.modal-body {
  padding: 20px 24px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.field-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--ink);
}
.semester-picker-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.picker-btn {
  flex: 1 1 140px;
  padding: 9px 12px;
  border-radius: 8px;
  border: 1px solid #dfe7e1;
  background: #f8fafc;
  color: var(--ink);
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
  text-align: center;
}
.picker-btn:hover {
  background: #eef3ef;
  border-color: var(--green);
}
.picker-btn.active {
  background: #eaf5ee;
  border-color: var(--green);
  color: #145c2d;
  box-shadow: 0 0 0 1px var(--green);
}
.custom-num-wrap {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
}
.custom-num-input {
  width: 60px;
  padding: 6px 8px;
  border: 1px solid #dfe7e1;
  border-radius: 6px;
  font-size: 11px;
}
.form-grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.field-input {
  width: 100%;
  padding: 9px 11px;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  background: #fff;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.field-input:focus {
  outline: none;
  border-color: var(--green);
  box-shadow: 0 0 0 3px rgba(76, 150, 112, 0.15);
}
.field-hint {
  font-size: 10px;
  color: var(--muted);
}
.modal-error-alert {
  padding: 9px 12px;
  border-radius: 8px;
  background: #fff0ec;
  border: 1px solid #f9cac2;
  color: #a8402f;
  font-size: 11px;
  font-weight: 600;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

/* Status colors */
.status-ongoing {
  border-color: #8bc99a;
  background: #dff4e4;
  color: #145c2d;
}
.status-coming {
  border-color: #e1bd5a;
  background: #fff3c4;
  color: #725300;
}
.status-ended {
  border-color: #bdc4c9;
  background: #edf0f2;
  color: #59636b;
}
.status-unknown {
  border-color: #d5dadd;
  background: #f5f6f7;
  color: #69737a;
}

/* Dark mode enhancements */
:global(html[data-theme="dark"]) .year-card-container {
  background: #1e293b !important;
  border-color: #334155;
  box-shadow: none;
}
:global(html[data-theme="dark"]) .year-card-active {
  border-color: #22c55e;
  box-shadow: 0 0 0 1px #22c55e;
}
:global(html[data-theme="dark"]) .stat-bubble,
:global(html[data-theme="dark"]) .btn-action-toggle {
  background: #0f172a;
  border-color: #334155;
  color: #f1f5f9 !important;
}
:global(html[data-theme="dark"]) .semester-section {
  background: #172033;
  border-top-color: #334155;
}
:global(html[data-theme="dark"]) .semester-count-chip {
  background: #0f172a;
  color: #86efac;
}
:global(html[data-theme="dark"]) .semester-card {
  background: #1e293b;
  border-color: #334155;
}
:global(html[data-theme="dark"]) .semester-card.is-current {
  border-color: #22c55e;
  background: linear-gradient(180deg, #132822 0%, #1e293b 45%);
}
:global(html[data-theme="dark"]) .semester-date-box,
:global(html[data-theme="dark"]) .semester-progress-card,
:global(html[data-theme="dark"]) .agenda-mini-chip {
  background: #0f172a;
  border-color: #334155;
}
:global(html[data-theme="dark"]) .duration-badge {
  background: #1e293b;
  border-color: #334155;
  color: #94a3b8;
}
:global(html[data-theme="dark"]) .progress-bar-bg {
  background: #334155;
}
:global(html[data-theme="dark"]) .card-action-btn {
  background: #0f172a;
  border-color: #334155;
  color: #f1f5f9 !important;
}
:global(html[data-theme="dark"]) .add-semester-card {
  background: #172033;
  border-color: #475569;
}
:global(html[data-theme="dark"]) .add-semester-card:hover {
  background: #132822;
  border-color: #22c55e;
}
:global(html[data-theme="dark"]) .add-card-circle {
  background: #1e293b;
  border-color: #475569;
}
:global(html[data-theme="dark"]) .empty-semesters-box {
  background: #1e293b;
  border-color: #334155;
}
:global(html[data-theme="dark"]) .modal-card {
  background: #1e293b;
  border-color: #334155;
}
:global(html[data-theme="dark"]) .modal-head {
  border-bottom-color: #334155;
}
:global(html[data-theme="dark"]) .picker-btn,
:global(html[data-theme="dark"]) .field-input,
:global(html[data-theme="dark"]) .custom-num-input {
  background: #0f172a;
  border-color: #334155;
  color: #f1f5f9;
}
:global(html[data-theme="dark"]) .picker-btn.active {
  background: #132822;
  border-color: #22c55e;
  color: #86efac;
}

/* Class distribution chips */
.class-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 10px;
}
.class-chip {
  padding: 6px 10px;
  border-radius: 999px;
  background: #f1f5f2;
  color: var(--ink);
  font-size: 10px;
  font-weight: 600;
}
.class-chip-warning {
  background: #fff0ec;
  color: #a8402f;
}
.empty-inline {
  margin-top: 8px;
  color: var(--muted);
  font-size: 11px;
}
.warning-item .item-mark {
  background: #fff0ec;
  color: #b45242;
}

/* Empty / loading states */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  padding: 32px 20px;
  text-align: left;
  color: var(--muted);
  font-size: 12px;
  border: 1px dashed #dfe7e1;
  border-radius: 10px;
}
.loading-state {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 32px 0;
  color: var(--muted);
  font-size: 12px;
}
.loader {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid #dfe7e1;
  border-top-color: var(--green);
  animation: spin 0.7s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Warnings */
.warning-text {
  grid-column: 1/-1;
  margin: 0;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fff7e5;
  color: #8b5a16;
  font-size: 11px;
  border: 1px solid #f4e1b0;
}
.reminder-badge {
  display: inline-block !important;
  margin-top: 10px;
  padding: 7px 10px;
  border-radius: 8px;
  background: #fff7e5;
  color: #8b5a16;
  font-size: 11px;
  font-weight: 600;
}
.agenda-list {
  display: grid;
  gap: 8px;
  margin-top: 12px;
}
.agenda-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 9px 10px;
  border: 1px solid #e5ebe6;
  border-radius: 8px;
  background: #fff;
  font-size: 11px;
}
.agenda-row strong {
  display: block;
  color: var(--ink);
}
.agenda-row span {
  display: block;
  margin-top: 3px;
  color: var(--muted);
  font-size: 10px;
}
.agenda-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.agenda-actions button {
  border: 0;
  background: transparent;
  color: var(--green);
  cursor: pointer;
  font-size: 10px;
  font-weight: 700;
}
.agenda-actions .link-action-danger {
  color: #b45242;
}
.agenda-form {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid var(--line);
}
.agenda-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 14px;
}
.agenda-form-grid label {
  display: grid;
  gap: 7px;
  color: var(--muted);
  font-size: 11px;
  font-weight: 600;
}
.agenda-form-grid > label {
  min-width: 0;
}
.agenda-form-grid input,
.agenda-form-grid select,
.agenda-form-grid textarea {
  width: 100%;
  border: 1px solid #dfe7e1;
  border-radius: 8px;
  padding: 10px 11px;
  background: #fff;
  color: var(--ink);
  font: inherit;
  font-size: 12px;
}
.date-time-fields {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(0, 0.65fr);
  gap: 7px;
}
.agenda-form-grid input:focus,
.agenda-form-grid select:focus,
.agenda-form-grid textarea:focus,
.form-grid input:focus,
.form-grid select:focus,
.form-grid textarea:focus {
  outline: none;
  border-color: var(--green);
  box-shadow: 0 0 0 3px rgba(76, 150, 112, 0.14);
}
.agenda-description {
  grid-column: 1/-1;
}
.overview-agenda {
  display: grid;
  gap: 6px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--line);
  font-size: 11px;
}
.overview-agenda span {
  color: var(--muted);
}

/* Timeline */
.timeline-panel {
  margin-top: 18px;
}
.calendar-toolbar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin: 18px 0 12px;
}
.calendar-toolbar h4 {
  min-width: 170px;
  margin: 0;
  text-align: center;
  text-transform: capitalize;
  font-size: 15px;
}
.calendar-nav,
.today-button {
  border: 1px solid #dfe7e1;
  border-radius: 7px;
  background: #fff;
  color: var(--ink);
  cursor: pointer;
  font: inherit;
}
.calendar-nav {
  width: 34px;
  height: 34px;
  font-size: 22px;
  line-height: 1;
}
.today-button {
  padding: 8px 11px;
  font-size: 11px;
  font-weight: 700;
}
.calendar-nav:hover,
.today-button:hover {
  border-color: var(--green);
  color: var(--green);
}
.calendar-weekdays,
.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
}
.calendar-weekdays {
  gap: 5px;
  margin-bottom: 5px;
}
.calendar-weekdays span {
  padding: 5px;
  text-align: center;
  color: var(--muted);
  font-size: 10px;
  font-weight: 700;
}
.calendar-grid {
  gap: 5px;
}
.calendar-day {
  position: relative;
  display: flex;
  min-height: 94px;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  border: 1px solid #e5ebe6;
  border-radius: 7px;
  padding: 7px;
  background: #fff;
  color: var(--ink);
  cursor: pointer;
  text-align: left;
  font: inherit;
}
.calendar-day:hover,
.calendar-day.selected {
  border-color: var(--green);
  box-shadow: 0 0 0 2px rgba(76, 150, 112, 0.12);
}
.calendar-day.muted {
  background: #f7f9f7;
  color: #aab4ad;
}
.calendar-day.today .calendar-day-number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--green);
  color: #fff;
}
.calendar-day-number {
  font-size: 11px;
  font-weight: 700;
}
.agenda-count {
  color: var(--muted);
  font-size: 9px;
}
.calendar-dots {
  display: flex;
  gap: 3px;
}
.calendar-dots i,
.calendar-legend i {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
}
.calendar-agenda-label {
  width: 100%;
  overflow: hidden;
  color: var(--ink);
  font-size: 9px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dot-red {
  background: #d65b52;
}
.dot-green {
  background: #4c9670;
}
.dot-blue {
  background: #4d86c5;
}
.dot-yellow {
  background: #d7a936;
}
.dot-purple {
  background: #8663ad;
}
.dot-gray {
  background: #8b969d;
}
.calendar-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 14px;
  margin: 12px 0;
  font-size: 10px;
  color: var(--muted);
}
.calendar-legend span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.agenda-detail-panel {
  margin: 16px 0;
  padding: 16px;
  border: 1px solid #dfe7e1;
  border-radius: 9px;
  background: #f8fbf8;
}
.agenda-detail-panel p {
  margin: 12px 0;
  font-size: 12px;
  line-height: 1.55;
}
.agenda-detail-panel small {
  display: block;
  margin-top: 7px;
  color: var(--muted);
  font-size: 10px;
}
.agenda-detail-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 12px;
}
.agenda-detail-meta span {
  padding: 5px 8px;
  border-radius: 999px;
  background: #e5f4e8;
  color: #34704b;
  font-size: 10px;
  font-weight: 700;
}
.agenda-detail-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 12px;
}
.agenda-detail-list button {
  border: 1px solid #dfe7e1;
  border-radius: 6px;
  padding: 6px 8px;
  background: #fff;
  color: var(--ink);
  cursor: pointer;
  font-size: 10px;
}
.agenda-detail-list button.active {
  border-color: var(--green);
  color: var(--green);
}
.agenda-row-button {
  display: block;
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  text-align: left;
}
.agenda-row-button strong,
.agenda-row-button span {
  display: block;
}
.timeline-year {
  display: grid;
  grid-template-columns: 180px 1fr;
  align-items: center;
  gap: 16px;
  padding: 14px 0;
  border-top: 1px solid var(--line);
}
.timeline-label {
  display: grid;
  gap: 4px;
  font-size: 12px;
}
.timeline-label span {
  color: var(--muted);
  font-size: 10px;
}
.timeline-track {
  min-width: 0;
}
.year-bar {
  position: relative;
  height: 34px;
  border-radius: 7px;
  background: #e5ebe6;
}
.year-bar.active {
  background: #d8eee0;
}
.year-bar.archived {
  background: #eceff0;
}
.semester-bar {
  position: absolute;
  top: 5px;
  bottom: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  border-radius: 5px;
  background: #4c9670;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
  cursor: default;
}
.year-bar.archived .semester-bar {
  background: #8b969d;
}

@media (max-width: 750px) {
  .toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
  .form-grid {
    grid-template-columns: 1fr;
  }
  .form-actions {
    grid-column: auto;
  }
  .year-item {
    flex-wrap: wrap;
  }
  .year-actions {
    width: 100%;
    margin-left: 0;
    margin-top: 12px;
    white-space: normal;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
  .semester-form {
    grid-template-columns: 1fr 1fr;
  }
  .semester-form-actions {
    grid-column: 1/-1;
    justify-content: flex-start;
  }
  .semester-row {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }
  .timeline-year {
    grid-template-columns: 1fr;
    gap: 8px;
  }
  .calendar-toolbar {
    flex-wrap: wrap;
  }
  .calendar-toolbar h4 {
    order: -1;
    width: 100%;
    flex-basis: 100%;
  }
  .calendar-day {
    min-height: 78px;
    padding: 5px;
  }
  .calendar-agenda-label {
    font-size: 8px;
  }
  .calendar-weekdays {
    gap: 2px;
  }
  .calendar-grid {
    gap: 2px;
  }
  .agenda-form-grid {
    grid-template-columns: 1fr;
  }
  .date-time-fields {
    grid-template-columns: 1fr 0.8fr;
  }
  .agenda-description {
    grid-column: auto;
  }
  .agenda-row {
    align-items: flex-start;
    flex-direction: column;
  }
  .agenda-actions {
    align-self: flex-end;
  }
}

@media print {
  .toolbar,
  .page-intro,
  .tabs,
  .form-panel,
  .item-actions,
  .panel-heading button,
  .toolbar-actions,
  .banner {
    display: none !important;
  }
  .timeline-panel {
    box-shadow: none;
    border: 0;
  }
  .timeline-year {
    break-inside: avoid;
  }
}

/* ================= DARK MODE STYLES ================= */
:global(html[data-theme='dark']) .tabs {
  background: #151f30 !important;
}
:global(html[data-theme='dark']) .tab-button {
  color: #94a3b8;
}
:global(html[data-theme='dark']) .tab-button:hover {
  background: #273549;
  color: #f1f5f9;
}
:global(html[data-theme='dark']) .tab-button.active {
  background: #1e293b !important;
  color: #65c9ab !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3) !important;
}
:global(html[data-theme='dark']) .close-button:hover {
  background: #273549 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .form-grid input,
:global(html[data-theme='dark']) .form-grid textarea,
:global(html[data-theme='dark']) .form-grid select,
:global(html[data-theme='dark']) .agenda-form-grid input,
:global(html[data-theme='dark']) .agenda-form-grid textarea,
:global(html[data-theme='dark']) .agenda-form-grid select {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .calendar-cell {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .agenda-row {
  border-color: #334155 !important;
}
</style>

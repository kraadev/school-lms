<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useApi } from '~/composables/useApi'
import { useAuthStore } from '~/stores/auth'
import {
  FileSpreadsheet, Plus, Trash2, Edit3, CheckCircle2, XCircle, Clock,
  Award, HelpCircle, Save, Check, X, ArrowLeft, Send, Sparkles, BookOpen,
  UserCheck, AlertCircle, FileText, ChevronRight, Download
} from 'lucide-vue-next'

definePageMeta({ middleware: 'auth' })
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { request } = useApi()

const quizId = route.params.id as string

// Dynamic state
const loading = ref(true)
const saving = ref(false)
const msg = ref('')
const isError = ref(false)
const quiz = ref<any>(null)
const questions = ref<any[]>([])
const activeTab = ref<'questions' | 'submissions'>('questions')

// Teacher: Submissions state
const attempts = ref<any[]>([])
const selectedAttempt = ref<any>(null)
const showGradeModal = ref(false)
const essayScores = ref<Record<string, number>>({})
const essayFeedbacks = ref<Record<string, string>>({})

// Teacher: Question modal state
const showQuestionModal = ref(false)
const editingQuestionId = ref<string | null>(null)
const questionForm = ref({
  question: '',
  type: 'single_choice',
  points: 10,
  order: 1,
  options: [
    { text: '', is_correct: true, order: 1 },
    { text: '', is_correct: false, order: 2 }
  ]
})

// Teacher: Excel Import state
const showImportModal = ref(false)
const importRawText = ref('')
const importError = ref('')

// Student state
const currentAttempt = ref<any>(null)
const studentAnswers = ref<Record<string, { option_id?: string; option_ids?: string[]; essay_answer?: string }>>({})
const studentEssayAnswers = ref<Record<string, string>>({})
const currentQuestionIndex = ref(0)
const quizResult = ref<any>(null)
const remainingSeconds = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

function setMsg(m: string, err = false) {
  msg.value = m
  isError.value = err
  setTimeout(() => { msg.value = '' }, 4000)
}

function idOf(val: any): string {
  if (!val) return ''
  if (typeof val === 'string') return val
  return val.id || val.ID || val._id || String(val)
}

function quizStatus(): string {
  return String(quiz.value?.status || quiz.value?.Status || '').toLowerCase()
}

const formattedRemainingTime = computed(() => {
  const minutes = Math.floor(remainingSeconds.value / 60)
  const seconds = remainingSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

const stopCountdown = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

const startCountdown = (attempt: any) => {
  stopCountdown()
  const deadline = attempt?.deadline_at || attempt?.DeadlineAt
  const startedAt = attempt?.started_at || attempt?.StartedAt
  const durationMinutes = Number(quiz.value?.duration || quiz.value?.Duration || 0)
  const deadlineTime = deadline
    ? new Date(deadline).getTime()
    : startedAt
      ? new Date(startedAt).getTime() + durationMinutes * 60 * 1000
      : Date.now() + durationMinutes * 60 * 1000

  const updateRemainingTime = () => {
    remainingSeconds.value = Math.max(0, Math.ceil((deadlineTime - Date.now()) / 1000))
  }

  updateRemainingTime()
  if (remainingSeconds.value <= 0) {
    void submitStudentQuiz(true)
    return
  }

  countdownTimer = setInterval(() => {
    updateRemainingTime()
    if (remainingSeconds.value <= 0) {
      stopCountdown()
      void submitStudentQuiz(true)
    }
  }, 1000)
}

// Data loaders
async function load() {
  loading.value = true
  try {
    if (auth.role === 'guru') {
      const qRes = await request(`/teacher/quizzes/${quizId}`)
      quiz.value = qRes?.data || qRes
      const questRes = await request(`/teacher/quizzes/${quizId}/questions`)
      questions.value = questRes?.data || questRes || []
    } else {
      const qRes = await request(`/student/quizzes/${quizId}`)
      quiz.value = qRes?.data || qRes
      questions.value = quiz.value?.questions || quiz.value?.Questions || []
    }
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal memuat kuis.', true)
  } finally {
    loading.value = false
  }
}

async function loadStudentAttempt() {
  if (auth.role !== 'siswa') return

  try {
    const res = await request(`/student/quizzes/${quizId}/attempts/latest`)
    currentAttempt.value = res?.data || res

    if (currentAttempt.value?.status === 'graded') {
      await loadResult()
    } else if (currentAttempt.value?.status === 'in_progress') {
      startCountdown(currentAttempt.value)
    }
  } catch {
    // Belum ada attempt: tampilkan tombol mulai kuis.
  }
}

async function loadAttempts() {
  try {
    const res = await request(`/teacher/quizzes/${quizId}/attempts`)
    attempts.value = res?.data || []
  } catch (e: any) {
    console.error('Error loading attempts:', e)
  }
}

async function publishCurrentQuiz() {
  if (questions.value.length === 0) {
    setMsg('Tambahkan minimal 1 soal sebelum mempublikasikan kuis.', true)
    return
  }
  saving.value = true
  try {
    await request(`/teacher/quizzes/${quizId}/publish`, { method: 'POST' })
    setMsg('Kuis berhasil dipublikasikan!')
    await load()
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal mempublikasikan kuis.', true)
  } finally {
    saving.value = false
  }
}

async function closeCurrentQuiz() {
  if (!confirm('Yakin ingin menutup kuis ini? Siswa tidak akan bisa mengerjakan lagi.')) return
  saving.value = true
  try {
    await request(`/teacher/quizzes/${quizId}/close`, { method: 'POST' })
    setMsg('Kuis berhasil ditutup.')
    await load()
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal menutup kuis.', true)
  } finally {
    saving.value = false
  }
}

async function deleteCurrentQuiz() {
  if (!confirm('Yakin ingin menghapus kuis ini? Semua soal dan hasil pengerjaan akan ikut terhapus.')) return
  saving.value = true
  try {
    await request(`/teacher/quizzes/${quizId}`, { method: 'DELETE' })
    setMsg('Kuis berhasil dihapus.')
    await router.push('/learning')
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal menghapus kuis.', true)
  } finally {
    saving.value = false
  }
}

// Teacher Question Handlers
function openAddQuestion() {
  editingQuestionId.value = null
  const nextOrder = questions.value.reduce((max, q) => Math.max(max, q.order || q.Order || 0), 0) + 1
  questionForm.value = {
    question: '',
    type: 'single_choice',
    points: 10,
    order: nextOrder,
    options: [
      { text: '', is_correct: true, order: 1 },
      { text: '', is_correct: false, order: 2 }
    ]
  }
  showQuestionModal.value = true
}

function addOptionForm() {
  questionForm.value.options.push({
    text: '',
    is_correct: false,
    order: questionForm.value.options.length + 1
  })
}

function removeOptionForm(idx: number) {
  if (questionForm.value.options.length > 2) {
    questionForm.value.options.splice(idx, 1)
  }
}

function selectCorrectOption(idx: number) {
  if (questionForm.value.type === 'single_choice' || questionForm.value.type === 'true_false') {
    questionForm.value.options.forEach((o, i) => { o.is_correct = (i === idx) })
  } else {
    questionForm.value.options[idx].is_correct = !questionForm.value.options[idx].is_correct
  }
}

async function saveQuestion() {
  if (!questionForm.value.question.trim()) {
    setMsg('Pertanyaan wajib diisi!', true)
    return
  }
  saving.value = true
  try {
    let qId = editingQuestionId.value
    if (!qId) {
      const nextOrder = questionForm.value.order || (questions.value.reduce((max, q) => Math.max(max, q.order || q.Order || 0), 0) + 1)
      const res = await request(`/teacher/quizzes/${quizId}/questions`, {
        method: 'POST',
        body: {
          question: questionForm.value.question,
          type: questionForm.value.type,
          points: questionForm.value.points,
          order: nextOrder
        }
      })
      const createdQ = res?.data || res
      qId = idOf(createdQ)
    } else {
      await request(`/teacher/questions/${qId}`, {
        method: 'PUT',
        body: {
          question: questionForm.value.question,
          type: questionForm.value.type,
          points: questionForm.value.points,
          order: questionForm.value.order
        }
      })
    }

    // Save options if not essay
    if (questionForm.value.type !== 'essay' && qId) {
      for (let oIdx = 0; oIdx < questionForm.value.options.length; oIdx++) {
        const opt = questionForm.value.options[oIdx]
        if (opt.text.trim()) {
          await request(`/teacher/questions/${qId}/options`, {
            method: 'POST',
            body: { text: opt.text, is_correct: Boolean(opt.is_correct), order: oIdx + 1 }
          })
        }
      }
    }

    setMsg('Soal berhasil disimpan!')
    showQuestionModal.value = false
    await load()
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal menyimpan soal.', true)
  } finally {
    saving.value = false
  }
}

async function deleteQuestion(q: any) {
  const qId = idOf(q)
  if (!qId) {
    setMsg('ID soal tidak valid.', true)
    return
  }
  if (!confirm('Yakin ingin menghapus soal ini?')) return
  try {
    await request(`/teacher/questions/${qId}`, { method: 'DELETE' })
    setMsg('Soal berhasil dihapus!')
    await load()
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal menghapus soal.', true)
  }
}

// Teacher Excel Import Handler
function parseCSVOrExcelText(text: string) {
  const lines = text.split('\n').filter(l => l.trim())
  const questionsList: any[] = []
  
  for (let i = 0; i < lines.length; i++) {
    const cols = lines[i].split('\t').length > 1 ? lines[i].split('\t') : lines[i].split(',')
    if (cols.length < 2) continue
    
    // Header check
    if (i === 0 && (cols[0].toLowerCase().includes('soal') || cols[0].toLowerCase().includes('question') || cols[0].toLowerCase().includes('no'))) {
      continue
    }

    const qText = cols[0]?.trim() || cols[1]?.trim()
    if (!qText) continue

    const qTypeRaw = (cols[1] || cols[2] || 'single_choice').toLowerCase().trim()
    let qType = 'single_choice'
    if (qTypeRaw.includes('essay') || qTypeRaw.includes('uraian')) qType = 'essay'
    else if (qTypeRaw.includes('jamak') || qTypeRaw.includes('multi')) qType = 'multiple_choice'

    const points = parseFloat(cols[2] || cols[3] || '10') || 10
    const rawOptions = (cols[3] || cols[4] || '').split('|')
    const rawAnswer = (cols[4] || cols[5] || 'A').toUpperCase().trim()

    const options: any[] = []
    if (qType !== 'essay') {
      rawOptions.forEach((optText: string, oIdx: number) => {
        const cleanOpt = optText.trim()
        if (cleanOpt) {
          const letter = String.fromCharCode(65 + oIdx)
          const isCorr = rawAnswer.includes(letter) || rawAnswer === String(oIdx + 1) || rawAnswer === cleanOpt.toUpperCase()
          options.push({ text: cleanOpt, is_correct: isCorr, order: oIdx + 1 })
        }
      })
    }

    questionsList.push({
      question: qText,
      type: qType,
      points: points,
      order: questionsList.length + 1,
      options: options
    })
  }

  return questionsList
}

async function handleImportSubmit() {
  if (!importRawText.value.trim()) {
    importError.value = 'Silahkan paste isi Excel/CSV terlebih dahulu!'
    return
  }
  importError.value = ''
  saving.value = true
  try {
    const parsed = parseCSVOrExcelText(importRawText.value)
    if (parsed.length === 0) {
      importError.value = 'Format data tidak terbaca. Pastikan format kolom sesuai petunjuk.'
      return
    }

    await request(`/teacher/quizzes/${quizId}/questions/bulk`, {
      method: 'POST',
      body: { questions: parsed }
    })

    setMsg(`Berhasil mengimpor ${parsed.length} soal!`)
    showImportModal.value = false
    importRawText.value = ''
    await load()
  } catch (e: any) {
    importError.value = e?.data?.message || 'Gagal mengimpor soal dari Excel.'
  } finally {
    saving.value = false
  }
}

function downloadTemplateExcel() {
  const csvContent = 'data:text/csv;charset=utf-8,' +
    'Soal,Tipe,Poin,Pilihan (pisahkan dengan |),Jawaban Benar\n' +
    'Berapa hasil dari 5 + 5?,single_choice,10,8 | 10 | 12 | 15,B\n' +
    'Manakah yang termasuk planet gas?,multiple_choice,10,Mars | Jupiter | Saturnus | Merkurius,B|C\n' +
    'Jelaskan tentang siklus air!,essay,20,,\n'
  const encodedUri = encodeURI(csvContent)
  const link = document.createElement('a')
  link.setAttribute('href', encodedUri)
  link.setAttribute('download', 'Template_Soal_Kuis.csv')
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// Teacher Grading Handlers
function answerIdOf(ans: any): string {
  const rawId = idOf(ans?.id ?? ans?.ID ?? ans?.answer_id ?? ans?.AnswerID)
  if (rawId && rawId !== '00000000-0000-0000-0000-000000000000') {
    return rawId
  }
  return idOf(ans?.question_id ?? ans?.QuestionID)
}

function answerQuestionText(ans: any): string {
  return ans?.question || ans?.Question || ans?.question_text || ans?.QuestionText || 'Pertanyaan tidak tersedia'
}

function answerTypeOf(ans: any): string {
  return ans?.type || ans?.Type || ans?.question_type || ans?.QuestionType || 'single_choice'
}

function normalizeAnswers(value: any): any[] {
  if (Array.isArray(value)) return value
  if (!value) return []
  if (Array.isArray(value.answers)) return value.answers
  if (Array.isArray(value.Answers)) return value.Answers
  if (Array.isArray(value.data)) return normalizeAnswers(value.data)
  return []
}

async function openAttemptDetail(att: any) {
  try {
    const res = await request(`/teacher/attempts/${idOf(att)}`)
    const payload = res?.data || res
    selectedAttempt.value = {
      ...payload,
      answers: normalizeAnswers(payload)
    }
    essayScores.value = {}
    essayFeedbacks.value = {}
    selectedAttempt.value.answers.forEach((ans: any) => {
      const ansId = answerIdOf(ans)
      if (!ansId) return
      if (answerTypeOf(ans) === 'essay') {
        essayScores.value[ansId] = Number(ans.score ?? 0)
        essayFeedbacks.value[ansId] = ans.feedback || ans.Feedback || ''
      }
    })
    showGradeModal.value = true
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal membuka detail jawaban.', true)
  }
}

async function saveEssayGrade(ans: any) {
  if (!selectedAttempt.value) return
  const ansId = idOf(ans)
  try {
    const score = essayScores.value[ansId] ?? 0
    const feedback = essayFeedbacks.value[ansId] || ''
    await request(`/teacher/attempts/${idOf(selectedAttempt.value)}/answers/${ansId}/grade`, {
      method: 'PATCH',
      body: { score: Number(score), feedback: feedback }
    })
    setMsg('Nilai essay berhasil disimpan!')
    await openAttemptDetail(selectedAttempt.value)
    await loadAttempts()
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal menyimpan nilai.', true)
  }
}

// Student Handlers
async function startStudentAttempt() {
  const attemptsUsed = Number(quiz.value?.attempts_used || quiz.value?.AttemptsUsed || 0)
  const maxAttempts = Number(quiz.value?.max_attempts || quiz.value?.MaxAttempts || 1)
  if (attemptsUsed >= maxAttempts) {
    setMsg('Kesempatan mengerjakan kuis sudah habis.', true)
    return
  }
  saving.value = true
  try {
    const res = await request(`/student/quizzes/${quizId}/start`, { method: 'POST' })
    currentAttempt.value = res?.data || res
    startCountdown(currentAttempt.value)
    setMsg('Kuis dimulai! Selamat mengerjakan.')
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal memulai kuis.', true)
  } finally {
    saving.value = false
  }
}

async function saveStudentAnswer(questionId: string, optionId?: string, essayText?: string, optionIds?: string[]) {
  if (!currentAttempt.value) return

  const normalizedOptionId = optionId?.trim?.() || undefined
  const normalizedOptionIds = (optionIds || []).filter(Boolean)

  const body: any = {}
  if (normalizedOptionIds.length > 0) {
    body.option_ids = normalizedOptionIds
  } else if (normalizedOptionId) {
    body.option_id = normalizedOptionId
  }

  if (essayText !== undefined) {
    body.essay_answer = essayText
  }

  if (Object.keys(body).length === 0) return

  try {
    await request(`/student/attempts/${idOf(currentAttempt.value)}/answers/${questionId}`, {
      method: 'PUT',
      body
    })
    if (essayText !== undefined) {
      studentEssayAnswers.value[questionId] = essayText
    }
  } catch (e: any) {
    console.error('Error saving answer:', e)
  }
}

function handleOptionChange(questionId: string, optionId: string, event: Event) {
  const target = event.target as HTMLInputElement
  const effectiveOptionId = (target.value || optionId || '').trim()
  const checked = target.checked
  const q = questions.value.find((q: any) => idOf(q) === questionId)
  const qType = q?.type || q?.Type

  if (!effectiveOptionId) return

  if (qType === 'multiple_choice') {
    const current = (studentAnswers.value[questionId]?.option_ids as string[]) || []
    let updated: string[]
    if (checked) {
      updated = Array.from(new Set([...current, effectiveOptionId]))
    } else {
      updated = current.filter((id: string) => id !== effectiveOptionId)
    }

    studentAnswers.value[questionId] = { option_ids: updated }

    if (updated.length === 0) return
    saveStudentAnswer(questionId, undefined, undefined, updated)
  } else {
    studentAnswers.value[questionId] = { option_id: effectiveOptionId }
    saveStudentAnswer(questionId, effectiveOptionId)
  }
}

async function submitStudentQuiz(autoSubmit = false) {
  if (!currentAttempt.value) return
  if (!autoSubmit && !confirm('Apakah Anda yakin ingin mengumpulkan kuis ini?')) return
  stopCountdown()
  saving.value = true
  try {
    const res = await request(`/student/attempts/${idOf(currentAttempt.value)}/submit`, { method: 'POST' })
    currentAttempt.value = res?.data || res
    setMsg(autoSubmit ? 'Waktu kuis telah habis. Jawaban dikumpulkan otomatis.' : 'Kuis berhasil dikumpulkan!')
    await loadResult()
  } catch (e: any) {
    setMsg(e?.data?.message || 'Gagal mengumpulkan kuis.', true)
  } finally {
    saving.value = false
  }
}

async function loadResult() {
  if (!currentAttempt.value) return
  try {
    const res = await request(`/student/attempts/${idOf(currentAttempt.value)}/result`)
    quizResult.value = res?.data || res
  } catch (e: any) {
    console.error('Error loading result:', e)
  }
}

onMounted(async () => {
  await load()
  if (auth.role === 'guru') {
    await loadAttempts()
  } else {
    await loadStudentAttempt()
  }
})

onBeforeUnmount(stopCountdown)
</script>

<template>
  <AppShell active="learning" eyebrow="Kelola Kuis" :title="quiz?.title || quiz?.Title || 'Kuis'">
    <div class="max-w-6xl mx-auto space-y-6 pb-12">

      <!-- Notification Toast -->
      <div v-if="msg" :class="[isError ? 'bg-red-50 text-red-700 border-red-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200', 'p-4 rounded-xl border flex items-center justify-between shadow-sm animate-fade-in']">
        <div class="flex items-center space-x-2">
          <AlertCircle v-if="isError" class="w-5 h-5 text-red-500" />
          <CheckCircle2 v-else class="w-5 h-5 text-emerald-500" />
          <span class="font-medium text-sm">{{ msg }}</span>
        </div>
        <button @click="msg = ''" class="text-slate-400 hover:text-slate-600"><X class="w-4 h-4" /></button>
      </div>

      <!-- Header Card -->
      <div class="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <button @click="router.push('/learning')" class="inline-flex items-center text-xs font-semibold text-slate-500 hover:text-primary-600 mb-2 transition-colors">
            <ArrowLeft class="w-4 h-4 mr-1" /> Kembali ke Ruang Belajar
          </button>
          <h1 class="text-2xl font-bold text-slate-900 tracking-tight">{{ quiz?.title || quiz?.Title || 'Detail Kuis' }}</h1>
          <p class="text-sm text-slate-500 mt-1 max-w-2xl">{{ quiz?.description || quiz?.Description || 'Tidak ada deskripsi.' }}</p>

          <div class="flex flex-wrap items-center gap-3 mt-4 text-xs font-medium text-slate-600">
            <span class="inline-flex items-center bg-slate-100 px-3 py-1 rounded-full">
              <Clock class="w-3.5 h-3.5 mr-1.5 text-primary-500" /> Durasi: {{ quiz?.duration || quiz?.Duration }} menit
            </span>
            <span class="inline-flex items-center bg-slate-100 px-3 py-1 rounded-full">
              <Award class="w-3.5 h-3.5 mr-1.5 text-amber-500" /> Max Percobaan: {{ quiz?.max_attempts || quiz?.MaxAttempts || 1 }}x
            </span>
            <span v-if="auth.role === 'guru'" :class="[quizStatus() === 'published' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-amber-50 text-amber-700 border-amber-200', 'px-3 py-1 rounded-full border capitalize font-semibold']">
              Status: {{ quizStatus() || 'Draft' }}
            </span>
          </div>
        </div>

        <!-- Teacher Action Buttons -->
        <div v-if="auth.role === 'guru'" class="flex flex-wrap gap-2">
          <AppButton @click="downloadTemplateExcel()" variant="secondary" size="sm">
            <Download class="w-4 h-4 mr-1.5" /> Template Excel
          </AppButton>
          <AppButton @click="showImportModal = true" variant="secondary" size="sm">
            <FileSpreadsheet class="w-4 h-4 mr-1.5 text-emerald-600" /> Impor Excel
          </AppButton>
          <AppButton @click="openAddQuestion()" size="sm">
            <Plus class="w-4 h-4 mr-1.5" /> Tambah Soal
          </AppButton>
          <AppButton v-if="quizStatus() === 'draft'" @click="publishCurrentQuiz()" variant="primary" size="sm" :loading="saving">
            <Send class="w-4 h-4 mr-1.5" /> Publikasikan Kuis
          </AppButton>
          <AppButton v-else-if="quizStatus() === 'published'" @click="closeCurrentQuiz()" variant="secondary" size="sm" :loading="saving">
            <XCircle class="w-4 h-4 mr-1.5 text-red-500" /> Tutup Kuis
          </AppButton>
          <AppButton @click="deleteCurrentQuiz()" variant="secondary" size="sm" :loading="saving" class="!border-red-200 !text-red-600 hover:!bg-red-50">
            <Trash2 class="w-4 h-4 mr-1.5" /> Hapus Kuis
          </AppButton>
        </div>
      </div>

      <!-- TEACHER VIEW: Tabs & Content -->
      <template v-if="auth.role === 'guru'">
        <!-- Tabs navigation -->
        <div class="flex space-x-1 border-b border-slate-200">
          <button @click="activeTab = 'questions'" :class="[activeTab === 'questions' ? 'border-primary-600 text-primary-600 font-bold' : 'border-transparent text-slate-500 hover:text-slate-700', 'py-3 px-6 border-b-2 text-sm transition-all flex items-center space-x-2']">
            <BookOpen class="w-4 h-4" />
            <span>Daftar Soal ({{ questions.length }})</span>
          </button>
          <button @click="activeTab = 'submissions'" :class="[activeTab === 'submissions' ? 'border-primary-600 text-primary-600 font-bold' : 'border-transparent text-slate-500 hover:text-slate-700', 'py-3 px-6 border-b-2 text-sm transition-all flex items-center space-x-2']">
            <UserCheck class="w-4 h-4" />
            <span>Hasil & Penilaian Siswa ({{ attempts.length }})</span>
          </button>
        </div>

        <!-- TAB 1: Questions List -->
        <div v-if="activeTab === 'questions'" class="space-y-4">
          <div v-if="questions.length === 0" class="bg-white rounded-2xl p-12 text-center border border-dashed border-slate-200">
            <HelpCircle class="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <h3 class="text-lg font-semibold text-slate-800">Belum Ada Soal</h3>
            <p class="text-sm text-slate-500 mt-1 max-w-md mx-auto">Klik tombol 'Tambah Soal' atau 'Impor Excel' di atas untuk membuat soal baru untuk kuis ini sebelum mempublikasikannya.</p>
          </div>

          <div v-for="(q, idx) in questions" :key="idOf(q) || idx" class="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm space-y-4 transition-all hover:border-slate-200">
            <div class="flex justify-between items-start">
              <div class="flex items-center space-x-3">
                <span class="w-8 h-8 rounded-lg bg-primary-50 text-primary-600 font-bold text-sm flex items-center justify-center">{{ q.order || q.Order || idx + 1 }}</span>
                <div>
                  <span class="text-xs font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-slate-100 text-slate-600 mr-2">
                    {{ (q.type || q.Type) === 'essay' ? 'Essay / Uraian' : (q.type || q.Type) === 'multiple_choice' ? 'Jawaban Jamak' : 'Pilihan Ganda' }}
                  </span>
                  <span class="text-xs font-semibold text-amber-600 bg-amber-50 px-2 py-0.5 rounded">{{ q.points || q.Points }} Poin</span>
                </div>
              </div>

              <div class="flex space-x-1">
                <button @click="deleteQuestion(q)" class="p-1.5 text-slate-400 hover:text-red-600 rounded-lg hover:bg-red-50 transition-colors" title="Hapus Soal"><Trash2 class="w-4 h-4" /></button>
              </div>
            </div>

            <p class="text-slate-800 font-medium text-base pl-11">{{ q.question || q.Question }}</p>

            <!-- Options List if choice -->
            <div v-if="(q.options || q.Options) && (q.options || q.Options).length > 0" class="pl-11 space-y-2">
              <div v-for="(opt, oIdx) in (q.options || q.Options)" :key="idOf(opt) || oIdx" :class="[(opt.is_correct ?? opt.IsCorrect) ? 'bg-emerald-50/70 border-emerald-200 text-emerald-900 font-medium' : 'bg-slate-50 border-slate-100 text-slate-700', 'p-3 rounded-xl border text-sm flex items-center justify-between']">
                <div class="flex items-center space-x-2.5">
                  <span class="w-6 h-6 rounded-full bg-white border border-slate-200 text-xs font-bold text-slate-600 flex items-center justify-center shadow-xs">
                    {{ String.fromCharCode(65 + oIdx) }}
                  </span>
                  <span>{{ opt.text || opt.Text }}</span>
                </div>
                <span v-if="opt.is_correct ?? opt.IsCorrect" class="inline-flex items-center text-xs font-bold text-emerald-600 bg-white px-2.5 py-1 rounded-lg border border-emerald-200 shadow-2xs">
                  <CheckCircle2 class="w-3.5 h-3.5 mr-1" /> Kunci Jawaban
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB 2: Submissions & Grading -->
        <div v-if="activeTab === 'submissions'" class="space-y-4">
          <div v-if="attempts.length === 0" class="bg-white rounded-2xl p-12 text-center border border-dashed border-slate-200">
            <UserCheck class="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <h3 class="text-lg font-semibold text-slate-800">Belum Ada Siswa Mengumpulkan</h3>
            <p class="text-sm text-slate-500 mt-1">Hasil pengerjaan siswa akan langsung muncul di halaman ini.</p>
          </div>

          <div v-else class="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden">
            <table class="w-full text-left text-sm text-slate-600">
              <thead class="bg-slate-50 text-slate-700 font-semibold border-b border-slate-100 text-xs uppercase tracking-wider">
                <tr>
                  <th class="py-4 px-6">Nama Siswa</th>
                  <th class="py-4 px-6">Percobaan Ke</th>
                  <th class="py-4 px-6">Status</th>
                  <th class="py-4 px-6">Waktu Selesai</th>
                  <th class="py-4 px-6">Nilai Akhir</th>
                  <th class="py-4 px-6 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100">
                <tr v-for="att in attempts" :key="idOf(att)" class="hover:bg-slate-50/50 transition-colors">
                  <td class="py-4 px-6 font-semibold text-slate-900">{{ att.student_name || 'Siswa' }}</td>
                  <td class="py-4 px-6">Percobaan #{{ att.attempt_number }}</td>
                  <td class="py-4 px-6">
                    <span :class="[att.status === 'graded' ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : att.status === 'submitted' ? 'bg-blue-50 text-blue-700 border-blue-200' : 'bg-slate-100 text-slate-600 border-slate-200', 'px-2.5 py-1 rounded-full text-xs border font-medium capitalize']">
                      {{ att.status }}
                    </span>
                  </td>
                  <td class="py-4 px-6 text-xs text-slate-500">{{ att.submitted_at ? new Date(att.submitted_at).toLocaleString('id-ID') : '-' }}</td>
                  <td class="py-4 px-6 font-bold text-slate-900 text-base">{{ att.score != null ? att.score : '-' }}</td>
                  <td class="py-4 px-6 text-right">
                    <AppButton @click="openAttemptDetail(att)" size="xs" variant="secondary">
                      <FileText class="w-3.5 h-3.5 mr-1" /> Detail & Nilai
                    </AppButton>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>

      <!-- STUDENT VIEW: Quiz Execution & Results -->
      <template v-if="auth.role === 'siswa'">
        <!-- Start Quiz Card if not started -->
        <div v-if="!currentAttempt && !quizResult" class="bg-white rounded-2xl p-8 border border-slate-100 shadow-sm text-center max-w-2xl mx-auto space-y-6">
          <div class="w-16 h-16 rounded-2xl bg-primary-50 text-primary-600 flex items-center justify-center mx-auto shadow-sm">
            <Sparkles class="w-8 h-8" />
          </div>
          <div>
            <h2 class="text-xl font-bold text-slate-900">Siap Mengerjakan Kuis?</h2>
            <p class="text-sm text-slate-500 mt-2">Pastikan koneksi internet Anda stabil. Setelah memulai, timer akan berjalan sesuai durasi kuis ({{ quiz?.duration }} menit).</p>
          </div>
          <AppButton
            @click="startStudentAttempt()"
            :loading="saving"
            :disabled="Number(quiz?.attempts_used || quiz?.AttemptsUsed || 0) >= Number(quiz?.max_attempts || quiz?.MaxAttempts || 1)"
            class="w-full sm:w-auto px-8 py-3"
          >
            <Sparkles class="w-4 h-4 mr-2" /> Mulai Pengerjaan
          </AppButton>
        </div>

        <!-- Quiz Question Interface (In Progress) -->
        <div v-if="currentAttempt && currentAttempt.status === 'in_progress'" class="space-y-6">
          <!-- Question Navigation Bar -->
          <div class="bg-white rounded-2xl p-4 border border-slate-100 shadow-sm flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-center space-x-2 overflow-x-auto py-1">
              <button v-for="(q, qIdx) in questions" :key="idOf(q) || qIdx" @click="currentQuestionIndex = qIdx" :class="[currentQuestionIndex === qIdx ? 'bg-primary-600 text-white font-bold' : (studentAnswers[idOf(q)]?.option_id || (studentAnswers[idOf(q)]?.option_ids && studentAnswers[idOf(q)]?.option_ids?.length) || studentEssayAnswers[idOf(q)]) ? 'bg-emerald-100 text-emerald-800 font-semibold' : 'bg-slate-100 text-slate-600 hover:bg-slate-200', 'w-9 h-9 rounded-xl text-xs flex items-center justify-center transition-all shadow-2xs']">
                {{ qIdx + 1 }}
              </button>
            </div>
            <div
              class="inline-flex items-center gap-2 rounded-xl border px-4 py-2 text-sm font-bold tabular-nums"
              :class="remainingSeconds <= 60 ? 'border-red-200 bg-red-50 text-red-700' : 'border-primary-100 bg-primary-50 text-primary-700'"
              aria-live="polite"
              aria-label="Sisa waktu pengerjaan"
            >
              <Clock class="w-4 h-4" />
              {{ formattedRemainingTime }}
            </div>
            <AppButton @click="submitStudentQuiz()" variant="primary" size="sm" :loading="saving">
              <Send class="w-4 h-4 mr-1.5" /> Kumpulkan Kuis
            </AppButton>
          </div>

          <!-- Current Question Card -->
          <div v-if="questions[currentQuestionIndex]" class="bg-white rounded-2xl p-8 border border-slate-100 shadow-sm space-y-6">
            <div class="flex justify-between items-center pb-4 border-b border-slate-100">
              <span class="text-sm font-bold text-slate-500">Soal Nomor {{ currentQuestionIndex + 1 }} dari {{ questions.length }}</span>
              <span class="text-xs font-semibold px-3 py-1 rounded-full bg-amber-50 text-amber-700 border border-amber-200">
                {{ questions[currentQuestionIndex].points || questions[currentQuestionIndex].Points }} Poin
              </span>
            </div>

            <p class="text-lg font-semibold text-slate-900 leading-relaxed">{{ questions[currentQuestionIndex].question || questions[currentQuestionIndex].Question }}</p>

            <!-- Options for Choice Question -->
            <div v-if="(questions[currentQuestionIndex].type || questions[currentQuestionIndex].Type) !== 'essay'" class="space-y-3 pt-2">
              <label v-for="(opt, oIdx) in (questions[currentQuestionIndex].options || questions[currentQuestionIndex].Options)" :key="idOf(opt) || oIdx" :class="[ (studentAnswers[idOf(questions[currentQuestionIndex])] && (Array.isArray(studentAnswers[idOf(questions[currentQuestionIndex])].option_ids) ? studentAnswers[idOf(questions[currentQuestionIndex])].option_ids.includes(idOf(opt)) : studentAnswers[idOf(questions[currentQuestionIndex])].option_id === idOf(opt)) ) ? 'border-primary-500 bg-primary-50/50 text-primary-900 shadow-xs' : 'border-slate-200 hover:border-slate-300 text-slate-700 bg-white', 'p-4 rounded-xl border flex items-center space-x-3 cursor-pointer transition-all']">
                <input
                  :type="(questions[currentQuestionIndex].type || questions[currentQuestionIndex].Type) === 'multiple_choice' ? 'checkbox' : 'radio'"
                  :name="'q_' + idOf(questions[currentQuestionIndex])"
                  :value="idOf(opt)"
                  :checked="(
                    (questions[currentQuestionIndex].type || questions[currentQuestionIndex].Type) === 'multiple_choice'
                      ? (studentAnswers[idOf(questions[currentQuestionIndex])] && Array.isArray(studentAnswers[idOf(questions[currentQuestionIndex])].option_ids) && studentAnswers[idOf(questions[currentQuestionIndex])].option_ids.includes(idOf(opt)))
                      : (studentAnswers[idOf(questions[currentQuestionIndex])]?.option_id === idOf(opt))
                  )"
                  @change="handleOptionChange(idOf(questions[currentQuestionIndex]), idOf(opt), $event)"
                  class="w-4 h-4 text-primary-600 border-slate-300 focus:ring-primary-500" />
                <span class="w-6 h-6 rounded-full bg-slate-100 text-slate-600 font-bold text-xs flex items-center justify-center mr-1">{{ String.fromCharCode(65 + oIdx) }}</span>
                <span class="font-medium text-sm">{{ opt.text || opt.Text }}</span>
              </label>
            </div>

            <!-- Textarea for Essay -->
            <div v-else class="space-y-2 pt-2">
              <label class="text-xs font-bold text-slate-600 uppercase tracking-wider">Tuliskan Jawaban Essay Anda:</label>
              <textarea v-model="studentEssayAnswers[idOf(questions[currentQuestionIndex])]" @blur="saveStudentAnswer(idOf(questions[currentQuestionIndex]), undefined, studentEssayAnswers[idOf(questions[currentQuestionIndex])])" rows="6" class="w-full p-4 rounded-xl border border-slate-200 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 text-sm text-slate-800 placeholder-slate-400" placeholder="Ketik jawaban essay di sini..."></textarea>
            </div>

            <!-- Footer Next/Prev -->
            <div class="flex justify-between items-center pt-6 border-t border-slate-100">
              <AppButton @click="currentQuestionIndex--" :disabled="currentQuestionIndex === 0" variant="secondary" size="sm">Sebelumnya</AppButton>
              <AppButton @click="currentQuestionIndex++" :disabled="currentQuestionIndex === questions.length - 1" variant="secondary" size="sm">Selanjutnya</AppButton>
            </div>
          </div>
        </div>

        <div v-if="currentAttempt && currentAttempt.status === 'submitted' && !quizResult" class="bg-amber-50 rounded-2xl p-8 border border-amber-200 text-center max-w-2xl mx-auto space-y-3">
          <Clock class="w-10 h-10 text-amber-600 mx-auto" />
          <h2 class="text-xl font-bold text-amber-900">Jawaban sedang dinilai</h2>
          <p class="text-sm text-amber-800">Hasil kuis, jawaban benar atau salah, dan nilai akan tampil setelah guru selesai menilai semua soal essay.</p>
        </div>

        <div v-if="quizResult" class="space-y-6">
          <div class="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <div>
                <p class="text-xs font-bold uppercase tracking-wider text-emerald-600">Penilaian selesai</p>
                <h2 class="text-2xl font-bold text-slate-900 mt-1">Hasil Kuis</h2>
              </div>
              <div class="text-left sm:text-right">
                <p class="text-xs text-slate-500">Nilai akhir</p>
                <p class="text-3xl font-black text-primary-600">{{ quizResult.score ?? 0 }}</p>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3 mt-6 max-w-md">
              <div class="rounded-xl bg-emerald-50 border border-emerald-100 p-3">
                <p class="text-xs text-emerald-700">Jawaban benar</p>
                <p class="text-xl font-bold text-emerald-800">{{ quizResult.correct ?? 0 }} / {{ quizResult.total_questions ?? 0 }}</p>
              </div>
              <div class="rounded-xl bg-slate-50 border border-slate-100 p-3">
                <p class="text-xs text-slate-500">Terjawab</p>
                <p class="text-xl font-bold text-slate-800">{{ quizResult.answered ?? 0 }} / {{ quizResult.total_questions ?? 0 }}</p>
              </div>
            </div>
          </div>

          <div class="space-y-4">
            <div v-for="(answer, answerIndex) in quizResult.answers || []" :key="answer.id || answer.question_id" class="bg-white rounded-2xl p-6 border border-slate-100 shadow-sm space-y-3">
              <div class="flex justify-between items-start gap-3">
                <p class="text-sm font-bold text-slate-500">Soal {{ answerIndex + 1 }}</p>
                <span v-if="answer.type === 'essay'" class="text-xs font-semibold px-2.5 py-1 rounded-full" :class="answer.score != null ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'">
                  {{ answer.score != null ? `Nilai ${answer.score} / ${answer.points}` : 'Belum dinilai' }}
                </span>
                <span v-else :class="answer.is_correct ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'" class="text-xs font-semibold px-2.5 py-1 rounded-full">
                  {{ answer.is_correct ? 'Benar' : 'Salah' }}
                </span>
              </div>
              <p class="font-semibold text-slate-900">{{ answer.question }}</p>
              <p v-if="answer.type === 'essay'" class="text-sm text-slate-700 whitespace-pre-wrap bg-slate-50 rounded-xl p-4">{{ answer.essay_answer || 'Tidak ada jawaban.' }}</p>
              <p v-else class="text-sm text-slate-700">Jawaban Anda: <strong>{{ answer.option_texts?.join(', ') || answer.option_text || 'Tidak dijawab' }}</strong></p>
              <p v-if="answer.feedback" class="text-sm text-slate-600 bg-blue-50 border border-blue-100 rounded-xl p-3">Umpan balik guru: {{ answer.feedback }}</p>
            </div>
          </div>
        </div>
      </template>

      <!-- Excel Import Modal (Teacher) -->
      <div v-if="showImportModal" class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4 animate-fade-in">
        <div class="bg-white rounded-3xl p-6 md:p-8 max-w-2xl w-full shadow-2xl space-y-5 border border-slate-100">
          <div class="flex justify-between items-center border-b border-slate-100 pb-4">
            <div class="flex items-center space-x-2.5">
              <FileSpreadsheet class="w-6 h-6 text-emerald-600" />
              <h3 class="text-lg font-bold text-slate-900">Impor Soal dari Excel / CSV</h3>
            </div>
            <button @click="showImportModal = false" class="text-slate-400 hover:text-slate-600"><X class="w-5 h-5" /></button>
          </div>

          <div class="text-xs text-slate-600 space-y-2 bg-slate-50 p-4 rounded-2xl border border-slate-200/80">
            <p class="font-bold text-slate-800">Petunjuk Format Kolom (Copy & Paste dari Excel):</p>
            <p><strong>Kolom 1:</strong> Soal Pertanyaan</p>
            <p><strong>Kolom 2:</strong> Tipe (<code>single_choice</code>, <code>multiple_choice</code>, <code>essay</code>)</p>
            <p><strong>Kolom 3:</strong> Poin (contoh: 10)</p>
            <p><strong>Kolom 4:</strong> Pilihan Jawaban (pisahkan dengan karakter <code>|</code> contoh: <code>Opsi A | Opsi B | Opsi C</code>)</p>
            <p><strong>Kolom 5:</strong> Jawaban Benar (contoh: <code>B</code> atau <code>Opsi B</code>)</p>
          </div>

          <div class="space-y-2">
            <label class="text-xs font-bold text-slate-700 uppercase">Paste Isi File Excel / CSV di sini:</label>
            <textarea v-model="importRawText" rows="8" class="w-full p-3 font-mono text-xs border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-slate-50/50" placeholder="Paste tabel dari Excel di sini..."></textarea>
          </div>

          <p v-if="importError" class="text-xs font-semibold text-red-600 bg-red-50 p-3 rounded-xl border border-red-200">{{ importError }}</p>

          <div class="flex justify-end space-x-3 pt-2">
            <AppButton @click="showImportModal = false" variant="secondary" size="sm">Batal</AppButton>
            <AppButton @click="handleImportSubmit()" :loading="saving" size="sm">Proses Impor Soal</AppButton>
          </div>
        </div>
      </div>

      <!-- Add/Edit Question Modal (Teacher) -->
      <div v-if="showQuestionModal" class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4 animate-fade-in">
        <div class="bg-white rounded-3xl p-6 md:p-8 max-w-2xl w-full shadow-2xl space-y-6 border border-slate-100 max-h-[90vh] overflow-y-auto">
          <div class="flex justify-between items-center border-b border-slate-100 pb-4">
            <h3 class="text-lg font-bold text-slate-900">{{ editingQuestionId ? 'Edit Soal' : 'Tambah Soal Manual' }}</h3>
            <button @click="showQuestionModal = false" class="text-slate-400 hover:text-slate-600"><X class="w-5 h-5" /></button>
          </div>

          <div class="space-y-4">
            <div>
              <label class="text-xs font-bold text-slate-700 uppercase mb-1 block">Pertanyaan / Soal</label>
              <textarea v-model="questionForm.question" rows="3" class="w-full p-3 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500" placeholder="Ketikkan soal pertanyaan..."></textarea>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="text-xs font-bold text-slate-700 uppercase mb-1 block">Tipe Soal</label>
                <select v-model="questionForm.type" class="w-full p-3 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white">
                  <option value="single_choice">Pilihan Ganda (1 Jawaban Benar)</option>
                  <option value="multiple_choice">Jawaban Jamak (>1 Jawaban Benar)</option>
                  <option value="essay">Essay / Uraian</option>
                </select>
              </div>
              <div>
                <label class="text-xs font-bold text-slate-700 uppercase mb-1 block">Poin Soal</label>
                <input v-model.number="questionForm.points" type="number" min="1" class="w-full p-3 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
              </div>
            </div>

            <!-- Options configuration for choices -->
            <div v-if="questionForm.type !== 'essay'" class="space-y-3 pt-2">
              <div class="flex justify-between items-center">
                <label class="text-xs font-bold text-slate-700 uppercase">Pilihan Jawaban (Tandai yang Benar)</label>
                <button @click="addOptionForm()" class="text-xs font-bold text-primary-600 hover:text-primary-700 inline-flex items-center"><Plus class="w-3.5 h-3.5 mr-1" /> Tambah Opsi</button>
              </div>

              <div v-for="(opt, idx) in questionForm.options" :key="idx" class="flex items-center space-x-2">
                <button type="button" @click="selectCorrectOption(idx)" :class="[opt.is_correct ? 'bg-emerald-500 text-white' : 'bg-slate-100 text-slate-400 hover:bg-slate-200', 'w-8 h-8 rounded-lg flex items-center justify-center transition-colors shadow-2xs']">
                  <Check class="w-4 h-4" />
                </button>
                <input v-model="opt.text" type="text" class="flex-1 p-2.5 text-sm border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500" :placeholder="'Opsi ' + String.fromCharCode(65 + idx)" />
                <button type="button" @click="removeOptionForm(idx)" class="p-2 text-slate-400 hover:text-red-600"><Trash2 class="w-4 h-4" /></button>
              </div>
            </div>
          </div>

          <div class="flex justify-end space-x-3 pt-4 border-t border-slate-100">
            <AppButton @click="showQuestionModal = false" variant="secondary" size="sm">Batal</AppButton>
            <AppButton @click="saveQuestion()" :loading="saving" size="sm">Simpan Soal</AppButton>
          </div>
        </div>
      </div>

      <!-- Essay Grading Modal (Teacher) -->
      <div v-if="showGradeModal && selectedAttempt" class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4 animate-fade-in">
        <div class="bg-white rounded-3xl p-6 md:p-8 max-w-3xl w-full shadow-2xl space-y-6 border border-slate-100 max-h-[90vh] overflow-y-auto">
          <div class="flex justify-between items-center border-b border-slate-100 pb-4">
            <div>
              <h3 class="text-lg font-bold text-slate-900">Detail Jawaban & Penilaian Essay</h3>
              <p class="text-xs text-slate-500">Siswa: {{ selectedAttempt.student_name || selectedAttempt.Student?.FullName || 'Siswa' }} | Percobaan #{{ selectedAttempt.attempt_number ?? selectedAttempt.AttemptNumber ?? 1 }}</p>
            </div>
            <button @click="showGradeModal = false" class="text-slate-400 hover:text-slate-600"><X class="w-5 h-5" /></button>
          </div>

          <div class="space-y-6">
            <div v-if="!selectedAttempt?.answers?.length" class="p-5 rounded-2xl bg-amber-50 border border-amber-200 text-sm text-amber-800">
              Belum ada jawaban yang tersedia untuk attempt ini.
            </div>

            <div v-else v-for="(ans, aIdx) in selectedAttempt.answers" :key="answerIdOf(ans) || aIdx" class="p-5 rounded-2xl bg-slate-50 border border-slate-200/80 space-y-3">
              <div class="flex justify-between items-start">
                <span class="text-xs font-bold text-slate-600 uppercase">Soal #{{ aIdx + 1 }} ({{ answerTypeOf(ans) === 'essay' ? 'Essay' : answerTypeOf(ans) === 'multiple_choice' ? 'Jawaban Jamak' : 'Pilihan Ganda' }})</span>
                <span class="text-xs font-bold text-amber-700 bg-amber-50 px-2 py-0.5 rounded border border-amber-200">
                  Nilai: {{ ans.score != null ? ans.score : 'Belum Dinilai' }} / {{ ans.points || 10 }}
                </span>
              </div>

              <p class="text-sm font-semibold text-slate-900">{{ answerQuestionText(ans) }}</p>

              <!-- Choice response -->
              <div v-if="answerTypeOf(ans) !== 'essay'" class="text-xs text-slate-700 bg-white p-3 rounded-xl border border-slate-200 flex items-center justify-between">
                <span>
                  Jawaban Siswa:
                  <strong v-if="ans.option_texts && ans.option_texts.length > 0">{{ ans.option_texts.join(', ') }}</strong>
                  <strong v-else-if="ans.option_text">{{ ans.option_text }}</strong>
                  <strong v-else>Tidak Dijawab</strong>
                </span>
                <span :class="[ans.is_correct ? 'text-emerald-600 font-bold' : 'text-red-500 font-bold']">
                  {{ ans.is_correct ? '✓ Benar' : '✗ Salah' }}
                </span>
              </div>

              <!-- Essay response + Grading Form -->
              <div v-else class="space-y-3 pt-1">
                <div class="bg-white p-4 rounded-xl border border-slate-200 text-sm text-slate-800 space-y-1">
                  <span class="text-xs font-bold text-slate-400 uppercase block">Jawaban Essay Siswa:</span>
                  <p class="whitespace-pre-wrap">{{ ans.essay_answer || ans.EssayAnswer || 'Tidak ada jawaban.' }}</p>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-3 gap-3 bg-white p-4 rounded-xl border border-slate-200">
                  <div>
                    <label class="text-xs font-bold text-slate-700 block mb-1">Beri Nilai (Max {{ ans.points || 10 }})</label>
                    <input v-model.number="essayScores[answerIdOf(ans)]" type="number" min="0" :max="ans.points || 10" class="w-full p-2 text-sm border border-slate-200 rounded-lg focus:ring-2 focus:ring-primary-500" />
                  </div>
                  <div class="md:col-span-2">
                    <label class="text-xs font-bold text-slate-700 block mb-1">Catatan / Umpan Balik</label>
                    <div class="flex space-x-2">
                      <input v-model="essayFeedbacks[answerIdOf(ans)]" type="text" class="flex-1 p-2 text-sm border border-slate-200 rounded-lg focus:ring-2 focus:ring-primary-500" placeholder="Catatan untuk siswa..." />
                      <AppButton @click="saveEssayGrade(ans)" size="xs"><Save class="w-3.5 h-3.5 mr-1" /> Simpan</AppButton>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>
  </AppShell>
</template>

<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in {
  animation: fadeIn 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>

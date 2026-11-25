<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useApi } from '~/composables/useApi'
import { useAuthStore } from '~/stores/auth'

definePageMeta({ middleware: 'auth' })
const route = useRoute(); const auth = useAuthStore(); const { request } = useApi(); const assignmentId = route.params.id as string
const submissions = ref<any[]>([]); const loading = ref(true); const errorMessage = ref(''); const message = ref(''); const scores = ref<Record<string, number>>({}); const feedback = ref<Record<string, string>>({})
const idOf = (item: any) => item?.id || item?.ID
const nameOf = (item: any) => item.student_name || item.StudentName || item.student?.full_name || item.Student?.FullName || 'Siswa'
const load = async () => { try { const response = await request<{ data?: any[] }>(`/teacher/assignments/${assignmentId}/submissions`); submissions.value = response.data || [] } catch (error: any) { errorMessage.value = error?.data?.message || 'Pengumpulan belum dapat dimuat.' } finally { loading.value = false } }
const grade = async (submission: any) => { try { await request(`/teacher/submissions/${idOf(submission)}/grade`, { method: 'PATCH', body: { score: Number(scores.value[idOf(submission)]), feedback: feedback.value[idOf(submission)] || '' } }); message.value = `Nilai ${nameOf(submission)} berhasil disimpan.`; await load() } catch (error: any) { errorMessage.value = error?.data?.message || 'Nilai belum dapat disimpan.' } }
onMounted(load)
</script>
<template><AppShell active="learning" eyebrow="Submission grading" title="Penilaian tugas"><NuxtLink class="button button-secondary back-link" to="/learning">Kembali ke learning</NuxtLink><div v-if="errorMessage" class="form-error">{{ errorMessage }}</div><div v-if="message" class="success">{{ message }}</div><div v-if="loading" class="loading-state"><span class="loader" /> Memuat pengumpulan...</div><section v-else class="submission-list"><article v-for="submission in submissions" :key="idOf(submission)" class="panel submission"><div><p class="eyebrow">{{ submission.status || submission.Status || 'submitted' }}</p><h3>{{ nameOf(submission) }}</h3><p>{{ submission.submitted_at || submission.SubmittedAt || 'Belum ada waktu pengumpulan' }} · {{ submission.file_count || submission.FileCount || 0 }} file</p></div><form class="grade-form" @submit.prevent="grade(submission)"><label>Nilai<input v-model.number="scores[idOf(submission)]" type="number" min="0" max="100" required /></label><label>Feedback<textarea v-model="feedback[idOf(submission)]" rows="2" /></label><AppButton type="submit">Simpan nilai</AppButton></form></article><div v-if="!submissions.length" class="empty-state">Belum ada pengumpulan.</div></section></AppShell></template>
<style scoped>.back-link{align-self:flex-start;margin:-10px 0 18px}.submission-list{display:grid;gap:12px}.submission{display:grid;grid-template-columns:1fr 1.2fr;gap:22px}.submission h3{margin:7px 0}.submission p:last-child{color:var(--muted);font-size:11px}.grade-form{display:grid;grid-template-columns:110px 1fr auto;align-items:end;gap:10px}.grade-form input,.grade-form textarea{width:100%;border:1px solid #dfe7e1;border-radius:8px;padding:10px;font:inherit;font-size:12px}.success{margin-bottom:12px;padding:12px 14px;border-radius:10px;background:#e5f4e8;color:#34704b;font-size:11px}@media(max-width:700px){.submission{grid-template-columns:1fr}.grade-form{grid-template-columns:1fr}}
:global(html[data-theme='dark']) .submission { background: #1e293b !important; border-color: #334155 !important; }
:global(html[data-theme='dark']) .submission h3 { color: #f8fafc !important; }
:global(html[data-theme='dark']) .grade-form input, :global(html[data-theme='dark']) .grade-form textarea { background: #151f30 !important; border-color: #334155 !important; color: #f1f5f9 !important; }
</style>

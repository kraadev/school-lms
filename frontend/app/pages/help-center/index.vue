<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ArrowLeft, Check, CheckCircle2, ChevronDown, CircleHelp, Headphones, MessageCircle, Plus, Send, ShieldCheck, Trash2, UserRound, X } from 'lucide-vue-next'
import { useApi } from '~/composables/useApi'
import { useChat } from '~/composables/useChat'
import { useChatWebSocket } from '~/composables/useChatWebSocket'
import { useChatStore } from '~/stores/chat'
import { useAuthStore } from '~/stores/auth'

definePageMeta({ middleware: 'auth' })

const auth = useAuthStore()
const chatStore = useChatStore()
const { request } = useApi()
const { getMessages } = useChat()
const { connect, disconnect, connected, sendMessage } = useChatWebSocket()

const loading = ref(true)
const saving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const faqs = ref<any[]>([])
const conversations = ref<any[]>([])
const selectedConversationId = ref('')
const messages = computed(() => selectedConversationId.value ? chatStore.getMessages(selectedConversationId.value) : [])
const messageInput = ref('')
const expandedFaq = ref('')
const faqFormOpen = ref(false)
const editingFaqId = ref('')
const faqForm = ref({ question: '', answer: '', category: 'Umum', published: false })
const faqCategories = computed(() => [...new Set(faqs.value.map(item => item.category || 'Umum'))])

const isAdmin = computed(() => auth.role === 'admin')
const activeConversation = computed(() => conversations.value.find(item => item.id === selectedConversationId.value))
const displayName = (user: any) => user?.display_name || user?.DisplayName || user?.student?.full_name || user?.Student?.FullName || user?.teacher?.full_name || user?.Teacher?.FullName || user?.username || user?.Username || 'Pengguna'
const roleName = (user: any) => user?.role?.name || user?.Role?.Name || 'Pengguna'
const normalizeConversation = (item: any) => ({ ...item, id: item.id || item.ID, status: item.status || item.Status || 'waiting', assigned_admin_id: item.assigned_admin_id || item.AssignedAdminID })
const normalizeMessage = (item: any) => ({ ...item, id: item.id || item.ID, conversation_id: item.conversation_id || item.ConversationID, sender_id: item.sender_id || item.SenderID, content: item.content || item.Content, created_at: item.created_at || item.CreatedAt, sender: item.sender || item.Sender })

const load = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const [faqResponse, conversationResponse] = await Promise.all([
      request<{ data?: any[] }>('/chat/faqs'),
      request<{ data?: any[] }>('/chat/customer-service'),
    ])
    faqs.value = faqResponse.data || []
    conversations.value = (conversationResponse.data || []).map(normalizeConversation)
    if (!selectedConversationId.value && conversations.value[0]) await selectConversation(conversations.value[0])
  } catch (error: any) {
    errorMessage.value = error?.data?.message || 'Help Center belum dapat dimuat.'
  } finally { loading.value = false }
}

const selectConversation = async (conversation: any) => {
  selectedConversationId.value = conversation.id
  const response = await getMessages(conversation.id)
  chatStore.setMessages(conversation.id, response.data.map(normalizeMessage))
}

const createTicket = async () => {
  try {
    const response = await request<{ data?: any }>('/chat/customer-service', { method: 'POST' })
    const item = normalizeConversation(response.data)
    conversations.value = [item, ...conversations.value.filter(existing => existing.id !== item.id)]
    await selectConversation(item)
    successMessage.value = 'Permintaan Customer Service masuk ke antrean admin.'
  } catch (error: any) { errorMessage.value = error?.data?.message || 'Chat Customer Service belum dapat dibuat.' }
}

const claimConversation = async (conversation: any) => {
  try {
    const response = await request<{ data?: any }>(`/admin/customer-service/${conversation.id}/claim`, { method: 'POST' })
    const updated = normalizeConversation(response.data)
    conversations.value = conversations.value.map(item => item.id === updated.id ? updated : item)
    await selectConversation(updated)
  } catch (error: any) { errorMessage.value = error?.data?.message || 'Percakapan belum dapat diterima.' }
}

const closeConversation = async (conversation: any) => {
  try {
    await request(`/admin/customer-service/${conversation.id}/close`, { method: 'POST' })
    conversations.value = conversations.value.map(item => item.id === conversation.id ? { ...item, status: 'closed' } : item)
  } catch (error: any) { errorMessage.value = error?.data?.message || 'Percakapan belum dapat ditutup.' }
}

const submitMessage = () => {
  const content = messageInput.value.trim()
  if (!content || !selectedConversationId.value || activeConversation.value?.status === 'closed') return
  if (!sendMessage(selectedConversationId.value, content)) { errorMessage.value = 'Koneksi chat belum siap.'; return }
  messageInput.value = ''
}

const resetFaqForm = () => { editingFaqId.value = ''; faqForm.value = { question: '', answer: '', category: 'Umum', published: false }; faqFormOpen.value = true }
const editFaq = (item: any) => { editingFaqId.value = item.id || item.ID; faqForm.value = { question: item.question || item.Question || '', answer: item.answer || item.Answer || '', category: item.category || item.Category || 'Umum', published: item.published ?? item.Published ?? false }; faqFormOpen.value = true }
const saveFaq = async () => {
  saving.value = true; errorMessage.value = ''
  try {
    const path = editingFaqId.value ? `/admin/faqs/${editingFaqId.value}` : '/admin/faqs'
    const response = await request<{ data?: any }>(path, { method: editingFaqId.value ? 'PUT' : 'POST', body: faqForm.value })
    if (editingFaqId.value) faqs.value = faqs.value.map(item => (item.id || item.ID) === editingFaqId.value ? response.data : item)
    else faqs.value = [response.data, ...faqs.value]
    faqFormOpen.value = false; successMessage.value = 'FAQ berhasil disimpan.'
  } catch (error: any) { errorMessage.value = error?.data?.message || 'FAQ belum dapat disimpan.' }
  finally { saving.value = false }
}
const deleteFaq = async (item: any) => {
  const id = item.id || item.ID
  if (!confirm('Hapus FAQ ini?')) return
  try { await request(`/admin/faqs/${id}`, { method: 'DELETE' }); faqs.value = faqs.value.filter(row => (row.id || row.ID) !== id) }
  catch (error: any) { errorMessage.value = error?.data?.message || 'FAQ belum dapat dihapus.' }
}

watch(messages, () => {}, { deep: true })
onMounted(async () => { connect(); await load() })
onUnmounted(disconnect)
</script>

<template>
  <AppShell active="help-center" eyebrow="School Space" title="Customer Service">
    <main class="help-page">
      <section class="help-hero">
        <div class="hero-icon"><Headphones :size="24" /></div>
        <div>
          <p class="eyebrow">Pusat bantuan</p>
          <h1>Temukan jawaban, lalu lanjutkan belajar.</h1>
          <p>Gunakan FAQ untuk panduan cepat. Jika masih membutuhkan bantuan, kirim pesan langsung ke admin.</p>
        </div>
        <button v-if="!isAdmin" class="primary-action" type="button" @click="createTicket"><MessageCircle :size="16" /> Mulai Chat CS</button>
      </section>

      <p v-if="errorMessage" class="notice error">{{ errorMessage }}</p>
      <p v-if="successMessage" class="notice success"><CheckCircle2 :size="15" /> {{ successMessage }}</p>

      <div class="help-layout">
        <section class="faq-section">
          <div class="section-heading"><div><p class="eyebrow">FAQ</p><h2>Pertanyaan umum</h2></div><button v-if="isAdmin" class="secondary-action" type="button" @click="resetFaqForm"><Plus :size="15" /> Tambah FAQ</button></div>
          <div v-if="faqCategories.length" class="category-row"><span v-for="category in faqCategories" :key="category" class="category-pill">{{ category }}</span></div>
          <div v-if="loading" class="empty-copy">Memuat pusat bantuan...</div>
          <article v-for="faq in faqs" :key="faq.id || faq.ID" class="faq-item">
            <button class="faq-question" type="button" @click="expandedFaq = expandedFaq === (faq.id || faq.ID) ? '' : (faq.id || faq.ID)"><span><CircleHelp :size="17" /> {{ faq.question || faq.Question }}</span><ChevronDown :size="17" :class="{ rotated: expandedFaq === (faq.id || faq.ID) }" /></button>
            <div v-if="expandedFaq === (faq.id || faq.ID)" class="faq-answer"><p>{{ faq.answer || faq.Answer }}</p><div v-if="isAdmin" class="faq-actions"><button type="button" @click="editFaq(faq)">Edit</button><button type="button" @click="deleteFaq(faq)"><Trash2 :size="13" /> Hapus</button></div></div>
          </article>
          <div v-if="!loading && !faqs.length" class="empty-copy">Belum ada FAQ yang dipublikasikan.</div>
        </section>

        <aside class="cs-section">
          <div class="section-heading"><div><p class="eyebrow">Live Customer Service</p><h2>{{ isAdmin ? 'Antrean bantuan' : 'Percakapan saya' }}</h2></div><span class="connection-dot" :class="{ online: connected }">{{ connected ? 'Online' : 'Offline' }}</span></div>
          <div v-if="isAdmin" class="queue-list">
            <button v-for="conversation in conversations" :key="conversation.id" type="button" class="queue-item" :class="{ selected: selectedConversationId === conversation.id }" @click="selectConversation(conversation)"><span class="queue-avatar"><UserRound :size="15" /></span><span class="queue-copy"><strong>{{ displayName(conversation.participants?.find((item: any) => item.user)?.user) }}</strong><small>{{ roleName(conversation.participants?.find((item: any) => item.user)?.user) }} · {{ conversation.status }} · {{ conversation.assigned_admin_id ? 'Ditangani' : 'Menunggu admin' }}</small></span></button>
          </div>
          <div v-else class="user-ticket-list"><div v-if="!conversations.length" class="empty-copy">Belum ada percakapan CS aktif.</div><button v-for="conversation in conversations" :key="conversation.id" type="button" class="queue-item" :class="{ selected: selectedConversationId === conversation.id }" @click="selectConversation(conversation)"><span class="queue-avatar"><MessageCircle :size="15" /></span><span class="queue-copy"><strong>Customer Service</strong><small>Status: {{ conversation.status }}</small></span></button></div>

          <div v-if="activeConversation" class="conversation-box">
            <div class="conversation-head"><div><strong>Customer Service</strong><small>Status: {{ activeConversation.status }}</small></div><div v-if="isAdmin" class="head-actions"><button v-if="activeConversation.status === 'waiting'" type="button" @click="claimConversation(activeConversation)"><Check :size="14" /> Terima</button><button v-if="activeConversation.status !== 'closed'" type="button" @click="closeConversation(activeConversation)"><X :size="14" /> Tutup</button></div></div>
            <div class="message-list"><div v-for="message in messages" :key="message.id" class="message-line" :class="{ mine: message.sender_id === auth.user?.id }"><span>{{ displayName(message.sender) }}</span><p>{{ message.content }}</p></div><div v-if="!messages.length" class="empty-copy">Belum ada pesan. Sampaikan pertanyaan Anda.</div></div>
            <form v-if="activeConversation.status !== 'closed'" class="message-form" @submit.prevent="submitMessage"><input v-model="messageInput" placeholder="Tulis pesan ke admin..." /><button type="submit" title="Kirim"><Send :size="16" /></button></form>
          </div>
        </aside>
      </div>

      <div v-if="faqFormOpen" class="modal-backdrop"><form class="faq-modal" @submit.prevent="saveFaq"><div class="modal-heading"><h2>{{ editingFaqId ? 'Edit FAQ' : 'Tambah FAQ' }}</h2><button type="button" @click="faqFormOpen = false"><X :size="17" /></button></div><label>Pertanyaan<input v-model="faqForm.question" required /></label><label>Jawaban<textarea v-model="faqForm.answer" rows="5" required /></label><label>Kategori<input v-model="faqForm.category" placeholder="Umum, Akun, Tugas..." /></label><label class="checkbox-label"><input v-model="faqForm.published" type="checkbox" /> Publikasikan FAQ</label><div class="modal-actions"><button class="secondary-action" type="button" @click="faqFormOpen = false">Batal</button><button class="primary-action" type="submit" :disabled="saving">Simpan</button></div></form></div>
    </main>
  </AppShell>
</template>

<style scoped>
.help-page { display:grid; gap:18px; }
.help-hero { display:flex; align-items:center; gap:16px; padding:24px; border:1px solid #dbe9df; border-radius:24px; background:linear-gradient(120deg,#f0fbf4,#fff9f2); }
.hero-icon { display:flex; align-items:center; justify-content:center; width:52px; height:52px; border-radius:17px; color:#087f5b; background:#d9f6e6; flex:none; }
.help-hero h1 { margin:3px 0 4px; color:#183b2a; font-size:clamp(1.2rem,2vw,1.8rem); }
.help-hero p:not(.eyebrow) { margin:0; color:#637166; font-size:12px; line-height:1.6; }
.help-hero .primary-action { margin-left:auto; flex:none; }
.help-layout { display:grid; grid-template-columns:minmax(0,1.15fr) minmax(330px,.85fr); gap:18px; align-items:start; }
.faq-section,.cs-section { padding:20px; border:1px solid #e3ebe5; border-radius:20px; background:#fff; box-shadow:0 5px 20px rgba(36,70,49,.05); }
.section-heading { display:flex; align-items:flex-start; justify-content:space-between; gap:12px; margin-bottom:14px; }
.section-heading h2 { margin:2px 0 0; color:#1f3427; font-size:18px; }
.primary-action,.secondary-action,.faq-actions button,.head-actions button { display:inline-flex; align-items:center; justify-content:center; gap:6px; border:0; border-radius:10px; padding:9px 12px; cursor:pointer; font:inherit; font-size:11px; font-weight:800; }
.primary-action { color:#fff; background:#168b65; }
.secondary-action { color:#226045; background:#e8f6ed; }
.category-row { display:flex; flex-wrap:wrap; gap:6px; margin-bottom:12px; }
.category-pill { padding:5px 9px; border-radius:99px; color:#377154; background:#eff8f1; font-size:10px; font-weight:700; }
.faq-item { border-top:1px solid #edf1ed; }
.faq-question { display:flex; align-items:center; justify-content:space-between; width:100%; padding:14px 2px; border:0; background:transparent; color:#24392c; cursor:pointer; font:inherit; font-size:12px; font-weight:800; text-align:left; }
.faq-question span { display:flex; align-items:center; gap:9px; }
.faq-question svg { flex:none; color:#4ca87e; transition:transform .2s; }.faq-question .rotated { transform:rotate(180deg); }
.faq-answer { padding:0 10px 14px 28px; color:#66736a; font-size:12px; line-height:1.65; }.faq-answer p { margin:0; white-space:pre-line; }
.faq-actions { display:flex; gap:6px; margin-top:10px; }.faq-actions button { padding:5px 8px; color:#377154; background:#eff8f1; }.faq-actions button:last-child { color:#b14d43; background:#fff0ed; }
.connection-dot { padding:5px 8px; border-radius:99px; color:#9a6b24; background:#fff5df; font-size:10px; font-weight:800; }.connection-dot.online { color:#087f5b; background:#e1f8eb; }
.queue-list,.user-ticket-list { display:grid; gap:6px; max-height:260px; overflow:auto; }.queue-item { display:flex; align-items:center; gap:9px; width:100%; padding:10px; border:1px solid transparent; border-radius:12px; background:#f7faf7; cursor:pointer; text-align:left; }.queue-item.selected { border-color:#9bd5b3; background:#eefaf2; }.queue-avatar { display:flex; align-items:center; justify-content:center; width:30px; height:30px; border-radius:10px; color:#168b65; background:#dff5e7; }.queue-copy { display:grid; gap:2px; min-width:0; }.queue-copy strong,.queue-copy small { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }.queue-copy strong { color:#294436; font-size:11px; }.queue-copy small { color:#78857b; font-size:10px; text-transform:capitalize; }
.conversation-box { margin-top:14px; overflow:hidden; border:1px solid #e4ece5; border-radius:14px; }.conversation-head { display:flex; align-items:center; justify-content:space-between; gap:8px; padding:12px; background:#f5faf6; }.conversation-head div:first-child { display:grid; gap:3px; }.conversation-head strong { color:#294436; font-size:12px; }.conversation-head small { color:#7a887d; font-size:10px; }.head-actions { display:flex; gap:5px; }.head-actions button { padding:6px 8px; color:#377154; background:#e3f5e9; }.head-actions button:last-child { color:#ae554b; background:#fff0ed; }
.message-list { display:grid; gap:8px; min-height:150px; max-height:270px; overflow:auto; padding:12px; background:#fff; }.message-line { max-width:88%; }.message-line.mine { justify-self:end; text-align:right; }.message-line span { color:#7a887d; font-size:9px; }.message-line p { margin:3px 0 0; padding:8px 10px; border-radius:10px 10px 10px 2px; color:#3c4b42; background:#f0f5f1; font-size:11px; line-height:1.45; text-align:left; }.message-line.mine p { border-radius:10px 10px 2px 10px; color:#fff; background:#168b65; }.message-form { display:flex; gap:7px; padding:10px; border-top:1px solid #edf1ed; }.message-form input { flex:1; min-width:0; border:1px solid #dce7df; border-radius:9px; padding:9px 10px; font:inherit; font-size:11px; }.message-form button { display:flex; align-items:center; justify-content:center; width:34px; border:0; border-radius:9px; color:#fff; background:#168b65; cursor:pointer; }.empty-copy { padding:18px 4px; color:#87938a; font-size:11px; text-align:center; }.notice { display:flex; align-items:center; gap:7px; margin:0; padding:10px 12px; border-radius:10px; font-size:11px; }.notice.error { color:#a23f35; background:#fff0ed; }.notice.success { color:#227b4f; background:#eaf8ef; }
.modal-backdrop { position:fixed; inset:0; z-index:50; display:flex; align-items:center; justify-content:center; padding:18px; background:rgba(19,35,25,.4); }.faq-modal { display:grid; gap:13px; width:min(560px,100%); padding:22px; border-radius:20px; background:#fff; box-shadow:0 20px 70px rgba(0,0,0,.18); }.modal-heading { display:flex; align-items:center; justify-content:space-between; }.modal-heading h2 { margin:0; color:#203a29; font-size:18px; }.modal-heading button { border:0; background:transparent; color:#68776d; cursor:pointer; }.faq-modal label { display:grid; gap:6px; color:#607066; font-size:11px; font-weight:700; }.faq-modal input,.faq-modal textarea { width:100%; border:1px solid #dce7df; border-radius:9px; padding:10px; color:#24392c; font:inherit; font-size:12px; }.checkbox-label { display:flex!important; grid-template-columns:none!important; align-items:center; gap:8px; }.checkbox-label input { width:auto; }.modal-actions { display:flex; justify-content:flex-end; gap:8px; }
@media (max-width:800px) { .help-hero { align-items:flex-start; flex-wrap:wrap; }.help-hero .primary-action { margin-left:68px; }.help-layout { grid-template-columns:1fr; } }

/* ================= DARK MODE STYLES ================= */
:global(html[data-theme='dark']) .help-hero {
  background: linear-gradient(135deg, #112d27 0%, #1e293b 100%) !important;
  border-color: #1f4a40 !important;
}
:global(html[data-theme='dark']) .help-hero h1 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .help-hero p:not(.eyebrow) {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .hero-icon {
  background: #1e3a32 !important;
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .faq-section,
:global(html[data-theme='dark']) .cs-section {
  background: #1e293b !important;
  border-color: #334155 !important;
  box-shadow: 0 5px 20px rgba(0, 0, 0, 0.25) !important;
}
:global(html[data-theme='dark']) .section-heading h2 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .category-pill {
  background: #273549 !important;
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .faq-item {
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .faq-question {
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .faq-question:hover {
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .faq-answer {
  color: #cbd5e1 !important;
}
:global(html[data-theme='dark']) .faq-actions button {
  background: #273549 !important;
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .faq-actions button:last-child {
  background: #3b1e22 !important;
  color: #fca5a5 !important;
}
:global(html[data-theme='dark']) .queue-item {
  background: #151f30 !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .queue-item.selected {
  background: #1e3a32 !important;
  border-color: #65c9ab !important;
}
:global(html[data-theme='dark']) .queue-avatar {
  background: #273549 !important;
  color: #65c9ab !important;
}
:global(html[data-theme='dark']) .queue-copy strong {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .queue-copy small {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .conversation-box {
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .conversation-head {
  background: #151f30 !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .conversation-head strong {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .conversation-head small {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .message-list {
  background: #0f172a !important;
}
:global(html[data-theme='dark']) .message-line p {
  background: #1e293b !important;
  color: #f1f5f9 !important;
  border: 1px solid #334155 !important;
}
:global(html[data-theme='dark']) .message-line.mine p {
  background: #168b65 !important;
  color: #ffffff !important;
  border-color: transparent !important;
}
:global(html[data-theme='dark']) .message-line span {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .message-form {
  background: #151f30 !important;
  border-color: #334155 !important;
}
:global(html[data-theme='dark']) .message-form input {
  background: #0f172a !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .faq-modal {
  background: #1e293b !important;
  border: 1px solid #334155 !important;
  box-shadow: 0 20px 70px rgba(0, 0, 0, 0.5) !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .modal-heading h2 {
  color: #f8fafc !important;
}
:global(html[data-theme='dark']) .faq-modal label {
  color: #94a3b8 !important;
}
:global(html[data-theme='dark']) .faq-modal input,
:global(html[data-theme='dark']) .faq-modal textarea {
  background: #151f30 !important;
  border-color: #334155 !important;
  color: #f1f5f9 !important;
}
:global(html[data-theme='dark']) .empty-copy {
  color: #94a3b8 !important;
}
</style>

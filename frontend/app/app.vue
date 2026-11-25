<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppSettings } from '~/composables/useAppSettings'

const { applyThemeToDOM } = useAppSettings()
onMounted(() => {
  applyThemeToDOM()
})
</script>

<template>
  <NuxtPage />
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700;800&family=Nunito:wght@400;500;600;700;800&display=swap');
@import 'tailwindcss';

:root { 
  --ink: #2b3641; 
  --muted: #8b959e; 
  --line: #eaedf1; 
  --paper: #f5f6f8; 
  --white: #ffffff;
  --teal: #65c9ab; 
  --teal-dark: #52a68c;
  --yellow: #ffbc0d; 
  --coral: #ff5a5f; 
  --coral-light: #ff7669;
}

html[data-theme='dark'] {
  --ink: #f1f5f9;
  --muted: #94a3b8;
  --line: #334155;
  --paper: #0f172a;
  --white: #1e293b;
}

html[data-theme='dark'] h1, 
html[data-theme='dark'] h2, 
html[data-theme='dark'] h3, 
html[data-theme='dark'] h4, 
html[data-theme='dark'] h5, 
html[data-theme='dark'] h6 {
  color: #f8fafc;
}

html[data-theme='dark'] .brand {
  color: #f8fafc;
}

html[data-theme='dark'] .search-bar input {
  background: #1e293b;
  border-color: #334155;
  color: #f1f5f9;
}

html[data-theme='dark'] .nav-item {
  color: #94a3b8;
}

html[data-theme='dark'] .nav-item:hover,
html[data-theme='dark'] .nav-item.active {
  background: #273549 !important;
  color: var(--teal) !important;
  box-shadow: none !important;
}

html[data-theme='dark'] .user-profile {
  background: #1e293b;
  border-color: #334155;
}

html[data-theme='dark'] .user-profile:hover {
  background: #273549;
}

html[data-theme='dark'] .card,
html[data-theme='dark'] .panel,
html[data-theme='dark'] .stat-card,
html[data-theme='dark'] .sidebar,
html[data-theme='dark'] .card-inner,
html[data-theme='dark'] .learning-item,
html[data-theme='dark'] .admin-data-card,
html[data-theme='dark'] .stat-mini-card,
html[data-theme='dark'] .record-card,
html[data-theme='dark'] .class-roster-card,
html[data-theme='dark'] .submission,
html[data-theme='dark'] .material-group,
html[data-theme='dark'] .faq-section,
html[data-theme='dark'] .cs-section,
html[data-theme='dark'] .settings-nav-card,
html[data-theme='dark'] .settings-content-area,
html[data-theme='dark'] .storage-card,
html[data-theme='dark'] .toggle-card,
html[data-theme='dark'] .theme-card,
html[data-theme='dark'] .pill-btn,
html[data-theme='dark'] .presence-card {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #f1f5f9;
}

html[data-theme='dark'] .bg-white {
  background-color: #1e293b !important;
  color: #f1f5f9;
}
html[data-theme='dark'] .bg-slate-50,
html[data-theme='dark'] .bg-gray-50,
html[data-theme='dark'] .bg-zinc-50 {
  background-color: #151f30 !important;
  color: #f1f5f9;
}
html[data-theme='dark'] .border-slate-200,
html[data-theme='dark'] .border-slate-100,
html[data-theme='dark'] .border-gray-200,
html[data-theme='dark'] .border-gray-100 {
  border-color: #334155 !important;
}
html[data-theme='dark'] .text-slate-900,
html[data-theme='dark'] .text-slate-800,
html[data-theme='dark'] .text-gray-900,
html[data-theme='dark'] .text-gray-800 {
  color: #f8fafc !important;
}
html[data-theme='dark'] .text-slate-700,
html[data-theme='dark'] .text-slate-600,
html[data-theme='dark'] .text-gray-700,
html[data-theme='dark'] .text-gray-600 {
  color: #cbd5e1 !important;
}
html[data-theme='dark'] .text-slate-500,
html[data-theme='dark'] .text-gray-500 {
  color: #94a3b8 !important;
}

html[data-theme='dark'] .setting-card-item,
html[data-theme='dark'] .session-card {
  background: #151f30;
  border-color: #334155;
}

html[data-theme='dark'] .profile-hub-banner {
  background: linear-gradient(135deg, #112d27 0%, #1e293b 100%);
  border-color: #1f4a40;
}

html[data-theme='dark'] .profile-hub-icon {
  background: #1e293b;
  border-color: #1f4a40;
}

html[data-theme='dark'] select,
html[data-theme='dark'] textarea,
html[data-theme='dark'] .text-input,
html[data-theme='dark'] .select-input {
  background: #151f30;
  border-color: #334155;
  color: #f1f5f9;
}

html[data-theme='dark'] .btn-outline,
html[data-theme='dark'] .button-secondary,
html[data-theme='dark'] .btn-action-outline {
  background: #1e293b;
  color: #f1f5f9;
  border-color: #334155;
}
html[data-theme='dark'] .btn-outline:hover,
html[data-theme='dark'] .button-secondary:hover,
html[data-theme='dark'] .btn-action-outline:hover {
  background: #273549;
}

html[data-theme='dark'] .course-item:hover,
html[data-theme='dark'] .tab-btn:hover {
  background: #273549 !important;
  color: #f1f5f9 !important;
}

html[data-theme='dark'] .tab-btn {
  color: #94a3b8 !important;
}

html[data-theme='dark'] .tab-btn.active {
  background: rgba(101, 201, 171, 0.18) !important;
  color: #65c9ab !important;
  border: 1px solid rgba(101, 201, 171, 0.3) !important;
}

html[data-theme='dark'] .tab-btn.active .tab-title,
html[data-theme='dark'] .tab-btn.active .tab-icon {
  color: #65c9ab !important;
}

html[data-theme='dark'] .tab-btn.active .tab-desc {
  color: #94a3b8 !important;
}

html[data-theme='dark'] .tab-btn.active:hover {
  background: rgba(101, 201, 171, 0.26) !important;
}

html[data-theme='dark'] .course-info strong,
html[data-theme='dark'] .stat-card strong,
html[data-theme='dark'] .panel-heading h3,
html[data-theme='dark'] .section-title,
html[data-theme='dark'] .section-heading,
html[data-theme='dark'] .item-text strong,
html[data-theme='dark'] .toggle-text strong,
html[data-theme='dark'] .inner-title,
html[data-theme='dark'] .theme-name,
html[data-theme='dark'] .storage-card-header strong,
html[data-theme='dark'] .diag-item strong {
  color: #f8fafc;
}

html[data-theme='dark'] .chat-preview-panel {
  background: #1e293b;
  border-color: #334155;
  color: #f1f5f9;
}
html[data-theme='dark'] .chat-preview-header {
  border-color: #334155;
}
html[data-theme='dark'] .chat-preview-item:hover {
  background: #273549;
}

/* Chat Module Dark Mode Defaults */
html[data-theme='dark'] .chat-page-bg {
  background-color: #0f172a !important;
}
html[data-theme='dark'] .chat-container-box {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
html[data-theme='dark'] .chat-sidebar {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
html[data-theme='dark'] .chat-sidebar-header {
  border-color: #334155 !important;
}
html[data-theme='dark'] .chat-topbar,
html[data-theme='dark'] .chat-footer-bar {
  background-color: #1e293b !important;
  border-color: #334155 !important;
}
html[data-theme='dark'] .chat-main-area {
  background-color: #0f172a !important;
}
html[data-theme='dark'] .chat-bubble-incoming {
  background-color: #1e293b !important;
  color: #f1f5f9 !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3) !important;
}
html[data-theme='dark'] .chat-bubble-incoming p {
  color: #f1f5f9 !important;
}
html[data-theme='dark'] .chat-input-wrapper {
  background-color: #0f172a !important;
  border-color: #334155 !important;
}
html[data-theme='dark'] .chat-input-textarea {
  color: #f1f5f9 !important;
}

/* Font scaling */
html[data-font-size='sm'] { font-size: 13.5px; }
html[data-font-size='md'] { font-size: 15px; }
html[data-font-size='lg'] { font-size: 16.5px; }

/* Density */
html[data-density='compact'] .content-area { padding-top: 18px; padding-bottom: 18px; }
html[data-density='compact'] .card, html[data-density='compact'] .panel { padding: 16px; }

/* Reduced motion */
html[data-reduce-motion='true'] * {
  animation-duration: 0.001ms !important;
  animation-iteration-count: 1 !important;
  transition-duration: 0.001ms !important;
}

* { box-sizing: border-box; }
body { margin: 0; background: var(--paper); color: var(--ink); font-family: 'Nunito', sans-serif; }
button, input { font: inherit; }
button, a { -webkit-tap-highlight-color: transparent; }
a { color: inherit; text-decoration: none; }
h1, h2, h3, h4, h5, h6 { font-family: 'Poppins', sans-serif; color: #1a1e23; }

.app-frame { display: flex; min-height: 100vh; overflow: hidden; }

/* Sidebar */
.sidebar { width: 260px; flex: 0 0 260px; display: flex; flex-direction: column; padding: 30px 24px; background: var(--white); border-right: 1px solid var(--line); transition: transform 0.3s ease; z-index: 50; }
.brand { display: flex; align-items: center; gap: 12px; padding: 0 0 40px; font-size: 22px; font-weight: 800; color: #2b3641; font-family: 'Poppins', sans-serif; letter-spacing: -0.5px; }
.brand-logo { width: 38px; height: 38px; object-fit: contain; flex-shrink: 0; filter: drop-shadow(0 4px 10px rgba(59, 130, 246, 0.2)); }
.auth-logo { width: 56px; height: 56px; object-fit: contain; margin-bottom: 20px; filter: drop-shadow(0 6px 14px rgba(59, 130, 246, 0.25)); }
.brand-mark { display: grid; place-items: center; width: 32px; height: 32px; border-radius: 50%; background: var(--coral); position: relative; }
.brand-mark::after { content: ''; position: absolute; width: 16px; height: 16px; border-radius: 50%; background: var(--teal); right: -6px; bottom: -2px; border: 2px solid white; }
.sidebar-role { display: flex; align-items: center; gap: 8px; margin: 0 0 16px 12px; color: var(--muted); font-size: 11px; text-transform: uppercase; font-weight: 700; letter-spacing: 0.8px; }

.main-nav { display: grid; gap: 8px; }
.nav-item { display: flex; align-items: center; gap: 16px; width: 100%; padding: 14px 16px; border-radius: 12px; color: #6a7682; font-size: 14px; font-weight: 700; transition: 0.2s ease; }
.nav-item span { font-size: 18px; color: #a4b1bc; }
.nav-item:hover, .nav-item.active { background: #fdfdfd; color: var(--teal); box-shadow: 0 4px 14px rgba(101, 201, 171, 0.15); }
.nav-item.active span, .nav-item:hover span { color: var(--teal); }
.nav-badge { margin-left: auto; display: grid; place-items: center; padding: 2px 8px; border-radius: 20px; background: var(--coral); color: white; font-size: 10px; font-weight: 800; }

.sidebar-bottom { margin-top: auto; }
.join-course-card { background: #f55b4e; border-radius: 20px; padding: 16px 14px 18px; position: relative; text-align: center; overflow: hidden; margin-top: auto; box-shadow: 0 10px 25px rgba(245, 91, 78, 0.25); display: flex; flex-direction: column; align-items: center; }
.join-course-img { width: 100%; height: 145px; object-fit: cover; object-position: center top; border-radius: 14px; margin-bottom: 14px; pointer-events: none; }
.join-btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; background: white; color: #f55b4e; font-weight: 800; font-size: 13px; padding: 10px 22px; border-radius: 12px; border: none; cursor: pointer; transition: transform 0.2s, box-shadow 0.2s; box-shadow: 0 4px 10px rgba(0,0,0,0.08); width: 85%; }
.join-btn:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.12); }

/* Main Content */
.content-wrapper { flex: 1; display: flex; flex-direction: column; height: 100vh; overflow: hidden; }
.content-area { flex: 1; padding: 30px 40px; overflow-y: auto; overflow-x: hidden; }

/* Topbar */
.topbar { display: flex; align-items: center; justify-content: space-between; gap: 20px; margin-bottom: 24px; }
.search-bar { flex: 1; max-width: 450px; position: relative; }
.search-bar input { width: 100%; padding: 14px 20px 14px 45px; border-radius: 20px; border: 1px solid var(--line); background: white; font-size: 14px; outline: none; transition: border-color 0.2s; color: var(--ink); }
.search-bar input:focus { border-color: var(--teal); box-shadow: 0 0 0 3px rgba(101, 201, 171, 0.1); }
.search-bar .icon { position: absolute; left: 16px; top: 50%; transform: translateY(-50%); color: var(--muted); font-size: 16px; }

.top-actions { display: flex; align-items: center; gap: 24px; margin-left: auto; }
.icon-btn { position: relative; display: inline-flex; align-items: center; justify-content: center; background: none; border: none; font-size: 20px; color: var(--muted); cursor: pointer; transition: color 0.2s; text-decoration: none; }
.icon-btn:hover { color: var(--ink); }
.user-profile { display: flex; align-items: center; gap: 12px; background: white; padding: 6px 16px 6px 6px; border-radius: 30px; border: 1px solid var(--line); cursor: pointer; transition: box-shadow 0.2s, background-color 0.2s; text-decoration: none; }
.user-profile:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.06); background: #fbfcfb; }
.avatar { display: grid; place-items: center; width: 36px; height: 36px; border-radius: 50%; background: var(--yellow); color: white; font-size: 16px; font-weight: 800; overflow: hidden; }
.avatar img { width: 100%; height: 100%; object-fit: cover; }
.user-info { display: flex; flex-direction: column; }
.user-name { font-size: 13px; font-weight: 800; color: var(--ink); font-family: 'Poppins', sans-serif; }
.user-handle { font-size: 10px; color: var(--muted); }

/* Common Components */
.section-title { font-size: 20px; font-weight: 700; margin: 0 0 20px; color: var(--ink); display: flex; justify-content: space-between; align-items: center; }
.view-all { font-size: 13px; color: var(--muted); font-weight: 600; text-decoration: underline; text-underline-offset: 4px; }
.view-all:hover { color: var(--ink); }

.card { background: white; border-radius: 20px; padding: 24px; border: 1px solid var(--line); }

.loading-state, .empty-state { display: grid; place-items: center; min-height: 180px; color: var(--muted); font-size: 14px; } 
.loader { width: 24px; height: 24px; margin-bottom: 12px; border: 3px solid #eaedf1; border-top-color: var(--teal); border-radius: 50%; animation: spin 1s linear infinite; } 
@keyframes spin { to { transform: rotate(360deg); } }

/* Mobile Hamburger */
.mobile-toggle { display: none; background: white; border: 1px solid var(--line); border-radius: 10px; padding: 10px; cursor: pointer; font-size: 20px; color: var(--ink); }
.mobile-overlay { display: none; position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 40; opacity: 0; transition: opacity 0.3s; pointer-events: none; }
.mobile-overlay.active { opacity: 1; pointer-events: auto; }

/* Dashboard Specifics — 3D pop-out hero */
/* Wrapper: provides the extra top space that the overflowing image needs */
.hero-wrapper { position: relative; margin-bottom: 30px; padding-top: 28px; }
/* Card: overflow:hidden clips everything inside it, z-index:1 puts it below the overlay */
.hero-section { display: flex; align-items: center; background: linear-gradient(135deg, #5eb69d 0%, #3d9e86 100%); border-radius: 24px; padding: 38px 45px; padding-right: 400px; color: white; position: relative; overflow: hidden; box-shadow: 0 12px 30px rgba(94,182,157,0.35); z-index: 1; }
.hero-content { flex: 1; }
.hero-content h1 { font-size: 30px; font-weight: 800; color: white; margin: 0 0 14px; line-height: 1.25; }
.hero-content p { font-size: 15px; margin: 0 0 22px; opacity: 0.95; max-width: 440px; line-height: 1.5; }
/* Floating image: sibling of card, absolutely positioned, z-index higher than card */
/* Extends from above the card (-70px top) down to card bottom — creates the pop-out */
.hero-image-overlay { position: absolute; right: 30px; top: 0; bottom: -12px; width: 400px; z-index: 10; display: flex; align-items: flex-end; justify-content: center; pointer-events: none; }
.hero-image-overlay img { height: 100%; width: auto; max-height: 340px; object-fit: contain; object-position: bottom center; filter: drop-shadow(0 12px 28px rgba(0,0,0,0.22)); transition: transform 0.4s ease; }
.hero-image-overlay img:hover { transform: scale(1.03) translateY(-4px); }
.primary-btn { background: white; color: #358b73; padding: 12px 26px; border-radius: 12px; font-weight: 800; font-size: 14px; border: none; cursor: pointer; transition: transform 0.2s, box-shadow 0.2s; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.primary-btn:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.12); }
.hero-actions-row { display: flex; align-items: center; gap: 20px; flex-wrap: wrap; }
.hero-stats-group { display: flex; align-items: center; gap: 14px; flex-wrap: wrap; }
.hero-stat-pill { display: flex; align-items: center; gap: 10px; background: rgba(255,255,255,0.22); backdrop-filter: blur(8px); padding: 6px 14px 6px 8px; border-radius: 30px; border: 1px solid rgba(255,255,255,0.3); }
.hero-stat-pill strong { font-size: 13px; font-weight: 800; color: white; display: block; line-height: 1.1; font-family: 'Poppins', sans-serif; }
.hero-stat-pill small { font-size: 10px; color: rgba(255,255,255,0.9); display: block; font-weight: 600; }
.pill-icon { width: 28px; height: 28px; border-radius: 50%; display: grid; place-items: center; color: white; }
.pill-icon.coral { background: var(--coral); }
.pill-icon.gold { background: var(--yellow); }

.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 30px; }
.stat-card { background: white; border-radius: 20px; padding: 24px; border: 1px solid var(--line); display: flex; align-items: center; gap: 16px; box-shadow: 0 4px 12px rgba(0,0,0,0.02); transition: transform 0.2s; }
.stat-card:hover { transform: translateY(-3px); }
.stat-icon { width: 50px; height: 50px; border-radius: 14px; display: grid; place-items: center; font-size: 24px; font-weight: bold; color: white; }
.stat-icon-blue { background: var(--teal); }
.stat-icon-coral { background: var(--coral); }
.stat-icon-gold { background: var(--yellow); }
.stat-icon-green { background: var(--teal-dark); }
.stat-card strong { display: block; font-size: 24px; font-weight: 800; color: var(--ink); line-height: 1.2; font-family: 'Poppins', sans-serif; }
.stat-card span { font-size: 13px; color: var(--muted); font-weight: 600; }

.dashboard-grid { display: grid; grid-template-columns: 2fr 1fr; gap: 24px; }
.panel { background: white; border-radius: 24px; padding: 28px; border: 1px solid var(--line); box-shadow: 0 4px 12px rgba(0,0,0,0.02); }
.panel-heading { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.panel-heading h3 { font-size: 18px; font-weight: 700; margin: 0; color: var(--ink); }
.text-link { font-size: 13px; color: var(--teal); font-weight: 700; text-decoration: none; }
.text-link:hover { text-decoration: underline; }

.course-list { display: flex; flex-direction: column; gap: 16px; }
.course-item { display: flex; align-items: center; gap: 16px; padding: 12px; border-radius: 16px; transition: background 0.2s; }
.course-item:hover { background: var(--paper); }
.course-icon { width: 48px; height: 48px; border-radius: 12px; background: var(--paper); border: 1px solid var(--line); }
.course-info { flex: 1; }
.course-info strong { display: block; font-size: 15px; font-weight: 700; color: var(--ink); margin-bottom: 4px; }
.course-info p { margin: 0; font-size: 13px; color: var(--muted); }
.row-date { font-size: 12px; color: var(--muted); font-weight: 600; white-space: nowrap; }

.quick-list { display: flex; flex-direction: column; gap: 12px; }
.quick-list a { display: flex; align-items: center; gap: 16px; padding: 16px; border-radius: 16px; border: 1px solid var(--line); transition: all 0.2s; text-decoration: none; }
.quick-list a:hover { border-color: var(--teal); background: rgba(101, 201, 171, 0.05); }
.quick-list .icon { font-size: 20px; color: var(--teal); width: 40px; height: 40px; background: rgba(101, 201, 171, 0.1); border-radius: 10px; display: grid; place-items: center; }
.quick-list strong { display: block; font-size: 14px; font-weight: 700; color: var(--ink); margin-bottom: 4px; }
.quick-list small { display: block; font-size: 12px; color: var(--muted); }

/* Existing small utilities */
.alert { padding: 14px 18px; border-radius: 12px; background: #fff0ec; color: #bd5845; font-size: 13px; font-weight: 600; margin-bottom: 20px; }
.button { border: 0; border-radius: 10px; padding: 12px 18px; cursor: pointer; font-size: 13px; font-weight: 700; font-family: 'Nunito', sans-serif; transition: transform 0.1s, opacity 0.2s; } 
.button-primary { background: var(--teal); color: white; } 
.button-primary:hover { opacity: 0.9; transform: translateY(-1px); }
.button-secondary { border: 1px solid var(--line); background: white; color: var(--ink); } 
.button-secondary:hover { background: var(--paper); }
.button-wide { width: 100%; text-align: center; } 
.button:disabled { opacity: .6; cursor: not-allowed; transform: none; }

@media (max-width: 1024px) {
  .sidebar { position: fixed; left: -260px; height: 100vh; box-shadow: 4px 0 20px rgba(0,0,0,0.05); }
  .sidebar.open { transform: translateX(260px); }
  .mobile-toggle { display: block; }
  .mobile-overlay { display: block; }
  .content-area { padding: 20px; }
  .search-bar { max-width: 250px; }
  .top-actions { gap: 15px; }
  .stat-grid { grid-template-columns: repeat(2, 1fr); }
  .hero-image { flex: 0 0 220px; max-width: 220px; }
}
@media (max-width: 820px) {
  .dashboard-grid { grid-template-columns: 1fr; }
}
@media (max-width: 640px) {
  .sidebar {
    top: auto;
    right: 0;
    bottom: 0;
    left: 0;
    width: 100%;
    height: auto;
    min-height: 72px;
    padding: 8px 8px calc(8px + env(safe-area-inset-bottom));
    border-top: 1px solid var(--line);
    border-right: 0;
    box-shadow: 0 -6px 20px rgba(0,0,0,0.08);
    transform: none;
  }
  .sidebar.open { transform: none; }
  .brand, .sidebar-role, .join-course-card { display: none; }
  .mobile-toggle, .mobile-overlay { display: none; }
  .main-nav {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(42px, 1fr));
    width: 100%;
    gap: 2px;
  }
  .nav-item {
    flex-direction: column;
    justify-content: center;
    gap: 4px;
    min-width: 0;
    padding: 6px 2px;
    border-radius: 10px;
    font-size: 9px;
    line-height: 1.1;
    text-align: center;
  }
  .nav-item .icon-nav { width: 19px; height: 19px; flex: 0 0 auto; }
  .content-area { padding-bottom: calc(92px + env(safe-area-inset-bottom)); }
  .search-bar { display: none; }
  .user-info { display: none; }
  .user-profile { padding: 4px; }
  .topbar { margin-bottom: 20px; }
  .stat-grid { grid-template-columns: 1fr; }
  .hero-wrapper { padding-top: 0; }
  .hero-section { flex-direction: column; padding: 30px 20px 180px; text-align: center; }
  .hero-image-overlay { right: 50%; bottom: -10px; transform: translateX(50%); top: auto; width: 260px; height: 200px; }
  .hero-content h1 { font-size: 24px; }
}
</style>
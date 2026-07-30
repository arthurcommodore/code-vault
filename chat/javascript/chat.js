import { apiJson } from "./api.js";

const CHAT_LOCAL_TTL_MS = 7 * 24 * 60 * 60 * 1000;

export default function chat(faqData) {
  return {
    faq: faqData || {},
    selectedCategory: null,
    messages: [],
    query: '',
    userInput: '',
    inputUnlocked: false,
    isSending: false,
    messageId: 0,
    isRestoring: false,
    persistTimer: null,

    init() {
      if (!Array.isArray(this.faq.categories)) {
        this.faq.categories = []
      }

      this.loadLocalState()

      this.$watch('messages', () => {
        this.scrollToBottom()
        this.queuePersistState()
      })

      this.$watch('inputUnlocked', () => {
        this.queuePersistState()
      })

      this.loadRemoteState()
    },

    selectCategory(cat) {
      this.selectedCategory = cat;
      this.query = '';
    },

    normalize(value) {
      return String(value || '')
        .toLowerCase()
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .trim()
    },

    allQuestions() {
      return (this.faq.categories || []).flatMap((cat) => {
        return (cat.questions || []).map((q) => ({
          ...q,
          category: cat.title
        }))
      })
    },

    filteredQuestions() {
      const source = this.selectedCategory
        ? (this.selectedCategory.questions || []).map((q) => ({
          ...q,
          category: this.selectedCategory.title
        }))
        : this.allQuestions()

      const query = this.normalize(this.query)
      if (!query) return source

      return source.filter((q) => {
        return this.normalize(`${q.category} ${q.q} ${q.a}`).includes(query)
      })
    },

    currentTitle() {
      if (this.selectedCategory) return this.selectedCategory.title
      return this.faq.all_categories || 'Todas'
    },

    addMessage(question, answer) {
      this.messages.push({
        id: this.messageId++,
        text: question,
        type: 'user',
        source: 'user',
        createdAt: this.nowISO()
      });

      this.messages.push({
        id: this.messageId++,
        text: answer,
        type: 'bot',
        source: 'faq',
        createdAt: this.nowISO()
      });

      this.queuePersistState()
      this.scrollToBottom()
    },

    async sendMessage() {
      if (!this.inputUnlocked || this.isSending || !this.userInput.trim()) return;

      const input = this.userInput.trim();
      const history = this.chatHistory();

      this.messages.push({
        id: this.messageId++,
        text: input,
        type: 'user',
        source: 'user',
        createdAt: this.nowISO()
      });

      const loadingId = this.messageId++;
      this.messages.push({
        id: loadingId,
        text: this.loadingMessage(),
        type: 'bot',
        source: 'ai',
        isLoading: true,
        createdAt: this.nowISO()
      });

      this.userInput = '';
      this.isSending = true;

      this.queuePersistState()
      this.scrollToBottom()

      try {
        const res = await apiJson('/api/chat/ia', {
          method: 'POST',
          timeout: 70000,
          body: {
            message: input,
            history: history
          }
        });

        this.updateMessage(loadingId, this.responseText(res), {
          source: 'ai',
          isLoading: false,
          supportValidation: this.responseValidation(res)
        });
      } catch (err) {
        console.error('Erro ao enviar mensagem para IA:', err);
        this.updateMessage(loadingId, this.errorMessage(), {
          source: 'ai',
          isLoading: false,
          supportValidation: this.emptySupportValidation()
        });
      } finally {
        this.isSending = false;
        this.scrollToBottom();
      }
    },

    requestAttendant() {
      this.unlockAIChat()
    },

    unlockAIChat() {
      if (this.inputUnlocked) return;

      this.inputUnlocked = true;

      this.messages.push({
        id: this.messageId++,
        text: this.unlockMessage(),
        type: 'bot',
        source: 'system',
        createdAt: this.nowISO()
      });

      this.queuePersistState()
      this.scrollToBottom()

      this.$nextTick(() => {
        const input = this.$el.querySelector('.chat-input input')
        if (input) input.focus()
      })
    },

    shouldShowAttendantOption() {
      const hasQuery = Boolean(this.query && this.query.trim())
      const noResults = this.filteredQuestions().length === 0
      return this.messages.length > 0 || this.selectedCategory !== null || hasQuery || noResults
    },

    resetChat() {
      this.messages = []
      this.userInput = ''
      this.inputUnlocked = false
      this.isSending = false
      this.selectedCategory = null
      this.query = ''
      this.messageId = 0
      this.clearPersistedState()
    },

    scrollToBottom() {
      this.$nextTick(() => {
        const el = this.$el.querySelector('.messages-area')
        if (!el) return
        el.scrollTop = el.scrollHeight
      })
    },

    inputPlaceholder() {
      return this.faq.fallback_message || '';
    },

    unlockMessage() {
      return this.faq.ai_unlocked_message || this.faq.attendant_unlocked_message || '';
    },

    aiChatButton() {
      return this.faq.ai_chat_button || this.faq.need_attendant || 'Conversar com a IA';
    },

    aiInputLockedHint() {
      return this.faq.ai_input_locked_hint || this.faq.input_locked_hint || '';
    },

    satisfactionLabel() {
      return this.faq.satisfaction_button || 'Atendimento satisfatório';
    },

    escalationLabel() {
      return this.faq.escalate_button || 'Encaminhar para chamado para um atendente';
    },

    escalationCreatingLabel() {
      return this.faq.escalation_creating || 'Criando solicitação...';
    },

    escalationCreatedLabel() {
      return this.faq.escalation_created || 'Solicitação de atendimento criada.';
    },

    escalationErrorLabel() {
      return this.faq.escalation_error || 'Não foi possível criar a solicitação agora.';
    },

    satisfactionThanks() {
      return this.faq.satisfaction_thanks || 'Obrigado pelo retorno.';
    },

    escalationNotice() {
      return this.faq.attendant_escalation_notice || '';
    },

    loadingMessage() {
      return this.faq.loading_message || 'Pensando...';
    },

    errorMessage() {
      return this.faq.ai_error_message || this.escalationNotice() || 'Não consegui responder agora.';
    },

    responseText(res) {
      if (!res?.ok) {
        return res?.data?.message || this.errorMessage();
      }

      return res?.data?.answer || res?.data?.message || this.errorMessage();
    },

    responseValidation(res) {
      if (!res?.ok) return this.emptySupportValidation();
      return this.normalizeSupportValidation(res?.data?.supportValidation);
    },

    emptySupportValidation() {
      return {
        showEscalation: false,
        satisfactionPercent: 70,
        humanSupportNeedPercent: 0,
        reason: ''
      };
    },

    normalizeSupportValidation(value) {
      return {
        showEscalation: Boolean(value?.showEscalation),
        satisfactionPercent: this.clampPercent(value?.satisfactionPercent ?? 70),
        humanSupportNeedPercent: this.clampPercent(value?.humanSupportNeedPercent ?? 0),
        reason: String(value?.reason || ''),
      };
    },

    clampPercent(value) {
      const number = Number(value);
      if (!Number.isFinite(number)) return 0;
      return Math.max(0, Math.min(100, Math.round(number)));
    },

    updateMessage(id, text, patch = {}) {
      const msg = this.messages.find((item) => item.id === id);
      if (!msg) return;
      msg.text = text;
      Object.assign(msg, patch);
      this.queuePersistState();
    },

    chatHistory() {
      return this.messages
        .filter((msg) => !msg.isLoading && msg.source !== 'system')
        .slice(-10)
        .map((msg) => ({
          role: msg.type === 'user' ? 'user' : 'assistant',
          content: msg.text
        }))
        .filter((msg) => msg.content && msg.content.trim());
    },

    messageSource(msg) {
      const source = String(msg?.source || '').toLowerCase();
      if (['user', 'faq', 'ai', 'system'].includes(source)) return source;
      return msg?.type === 'user' ? 'user' : 'faq';
    },

    feedbackValue(value) {
      const feedback = String(value || '').toLowerCase();
      if (['satisfied', 'attendant_requested'].includes(feedback)) return feedback;
      return '';
    },

    nowISO() {
      return new Date().toISOString();
    },

    storageKey() {
      const lang = window.Alpine?.store('configLang')?.getLang?.() || window.location.pathname.split('/')[1] || 'en';
      return `cotarpreco:chat:${lang}`;
    },

    currentState() {
      return {
        savedAt: this.nowISO(),
        messages: this.persistableMessages(),
        inputUnlocked: Boolean(this.inputUnlocked),
        messageId: Number(this.messageId) || 0
      };
    },

    persistableMessages() {
      return this.messages.slice(-80).map((msg) => ({
          id: Number(msg.id) || 0,
          text: String(msg.text || ''),
          type: msg.type === 'user' ? 'user' : 'bot',
          source: this.messageSource(msg),
          supportValidation: msg.supportValidation ? this.normalizeSupportValidation(msg.supportValidation) : null,
          feedback: this.feedbackValue(msg.feedback),
          humanSupportTicketId: String(msg.humanSupportTicketId || ''),
          humanSupportStatus: String(msg.humanSupportStatus || ''),
          humanSupportError: String(msg.humanSupportError || ''),
          createdAt: msg.createdAt || this.nowISO()
        }));
    },

    applyState(state) {
      if (!state || !Array.isArray(state.messages)) return;

      this.isRestoring = true;

      const messages = state.messages
        .map((msg, index) => ({
          id: Number.isFinite(Number(msg.id)) ? Number(msg.id) : index,
          text: String(msg.text || ''),
          type: msg.type === 'user' ? 'user' : 'bot',
          source: this.messageSource(msg),
          supportValidation: msg.supportValidation ? this.normalizeSupportValidation(msg.supportValidation) : null,
          feedback: this.feedbackValue(msg.feedback),
          humanSupportTicketId: String(msg.humanSupportTicketId || ''),
          humanSupportStatus: String(msg.humanSupportStatus || ''),
          humanSupportError: String(msg.humanSupportError || ''),
          isLoading: false,
          createdAt: msg.createdAt || this.nowISO()
        }))
        .filter((msg) => msg.text.trim());

      this.messages = messages;
      this.inputUnlocked = Boolean(state.inputUnlocked);

      const nextId = messages.reduce((max, msg) => Math.max(max, msg.id + 1), 0);
      this.messageId = Math.max(Number(state.messageId) || 0, nextId);

      this.$nextTick(() => {
        this.isRestoring = false;
        this.scrollToBottom();
      });
    },

    loadLocalState() {
      try {
        localStorage.removeItem(this.storageKey());
        const raw = sessionStorage.getItem(this.storageKey());
        if (!raw) return;

        const state = JSON.parse(raw);
        const savedAt = Date.parse(state.savedAt || '');
        if (!savedAt || Date.now() - savedAt > CHAT_LOCAL_TTL_MS) {
          sessionStorage.removeItem(this.storageKey());
          return;
        }

        this.applyState(state);
      } catch (err) {
        console.warn('Não foi possível restaurar histórico local do chat:', err);
      }
    },

    async loadRemoteState() {
      try {
        const res = await apiJson('/api/chat/history', { method: 'GET' });
        if (!res?.ok) return;
        this.applyState(res.data);
        this.saveLocalState();
      } catch (err) {
        console.warn('Não foi possível restaurar histórico remoto do chat:', err);
      }
    },

    saveLocalState() {
      try {
        sessionStorage.setItem(this.storageKey(), JSON.stringify(this.currentState()));
      } catch (err) {
        console.warn('Não foi possível salvar histórico local do chat:', err);
      }
    },

    queuePersistState() {
      if (this.isRestoring) return;

      this.saveLocalState();

      if (this.persistTimer) {
        clearTimeout(this.persistTimer);
      }

      this.persistTimer = setTimeout(() => {
        this.persistRemoteState();
      }, 350);
    },

    async persistRemoteState() {
      try {
        await apiJson('/api/chat/history', {
          method: 'PUT',
          body: this.currentState()
        });
      } catch (err) {
        console.warn('Não foi possível salvar histórico remoto do chat:', err);
      }
    },

    clearPersistedState() {
      try {
        sessionStorage.removeItem(this.storageKey());
        localStorage.removeItem(this.storageKey());
      } catch (err) {
        console.warn('Não foi possível limpar histórico local do chat:', err);
      }

      if (this.persistTimer) {
        clearTimeout(this.persistTimer);
        this.persistTimer = null;
      }

      apiJson('/api/chat/history', { method: 'DELETE' }).catch((err) => {
        console.warn('Não foi possível limpar histórico remoto do chat:', err);
      });
    },

    shouldShowSatisfactionActions(msg) {
      return msg?.type === 'bot'
        && msg?.source === 'ai'
        && !msg?.isLoading
        && !msg?.feedback;
    },

    shouldShowEscalationAction(msg) {
      return this.shouldShowSatisfactionActions(msg)
        && Boolean(msg?.supportValidation?.showEscalation);
    },

    isCreatingHumanSupport(msg) {
      return msg?.isCreatingHumanSupport === true;
    },

    markSatisfactory(msg) {
      if (!this.shouldShowSatisfactionActions(msg)) return;
      msg.feedback = 'satisfied';
      this.queuePersistState();
    },

    async requestHumanSupport(msg) {
      if (!this.shouldShowEscalationAction(msg) || this.isCreatingHumanSupport(msg)) return;

      msg.isCreatingHumanSupport = true;
      msg.humanSupportError = '';
      this.queuePersistState();

      try {
        const res = await apiJson('/api/chat/human-support', {
          method: 'POST',
          body: {
            messages: this.persistableMessages(),
            supportValidation: this.normalizeSupportValidation(msg.supportValidation)
          }
        });

        if (!res?.ok) {
          throw new Error(res?.data?.message || this.escalationErrorLabel());
        }

        msg.feedback = 'attendant_requested';
        msg.humanSupportTicketId = String(res?.data?.ticket?.id || '');
        msg.humanSupportStatus = res?.data?.message || this.escalationCreatedLabel();
      } catch (err) {
        console.error('Erro ao criar solicitação de atendimento:', err);
        msg.humanSupportError = err?.message || this.escalationErrorLabel();
      } finally {
        msg.isCreatingHumanSupport = false;
        this.queuePersistState();
      }
    },

    messageFeedbackText(msg) {
      if (msg?.feedback === 'satisfied') return this.satisfactionThanks();
      if (msg?.humanSupportError) return msg.humanSupportError;
      if (msg?.feedback === 'attendant_requested') return msg.humanSupportStatus || this.escalationCreatedLabel();
      return '';
    },

    formatMessage(value) {
      const escaped = this.escapeHtml(value);
      const linked = escaped.replace(
        /((https?:\/\/[^\s<]+)|\/[a-z]{2}\/app\/[a-z0-9/_?=&%.:-]*)/gi,
        (url) => {
          const href = url.replace(/&amp;/g, '&');
          const target = href.startsWith('http') ? ' target="_blank" rel="noopener noreferrer"' : '';
          return `<a href="${href}"${target}>${url}</a>`;
        }
      );

      return linked.replace(/\n/g, '<br>');
    },

    escapeHtml(value) {
      return String(value || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
    },

    email() {
      return this.faq.contact_email || '';
    }
  }
}

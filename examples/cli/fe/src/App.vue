<template>
  <div class="root">
    <div class="chat" tabindex="-1" @keydown="onInputKey">
      <div class="out" ref="scrollParent">
        <pre v-html="output"></pre>
        <span :class="{ title: true, show: outputEmpty }">{{ name }}</span>
      </div>
      <div class="head">{{ pwd }} #</div>
      <div class="in">
        <textarea autofocus v-model="input" ref="inputElement" />
      </div>
    </div>
    <div :class="{ plugin: true, show: showPlugins }">
      <vim
        @provideHook="injectHook"
        @escape="inputElement.focus()"
        :load="load"
        :save="save"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { getapi, Base } from "vue2go";
import _ from "underscore";
import {
  reactive,
  toRefs,
  onMounted,
  ref,
  computed
} from "@vue/composition-api";
import Vim from "./components/vim.vue";
import api from "./api";

export default {
  components: {
    vim: Vim
  },
  setup() {
    const state = reactive({
      name: "",
      input: "",
      output: ""
    });
    const scrollParent = ref(null);
    const inputElement = ref(null);
    const outputEmpty = computed(() => state.output === "");
    const bufferLimit = 1000 * 80; // output buffer char count limit

    useMock(state);

    const { autoscroll, scrollToStart, scrollToEnd } = useScrollUI(
      scrollParent
    );
    const { printText, inputFeedback, pwd } = usePrintText(
      state,
      bufferLimit,
      autoscroll,
      scrollToEnd
    );
    const { handleCommand, clear } = useTerminalCommand(state, scrollToStart);
    const { SIGINT, SIGKILL } = useSendSignal();
    const { back, forward, push } = useHistory(state);
    const { hookInput, injectHook, pluginStatus } = usePlugin(state);
    const { onInputKey } = useUserInput(
      state,
      SIGINT,
      SIGKILL,
      autoscroll,
      clear,
      printText,
      inputFeedback,
      back,
      forward,
      push,
      inputElement,
      hookInput
    );
    const showPlugins = computed(() => {
      for (const active of pluginStatus.value) {
        if (active.value === true) return true;
      }
      return false;
    });

    onMounted(async function() {
      // (window as any).chat = this as any;
      state.name = await api.name();
      try {
        await api.listen((s, fp) => {
          if (handleCommand(s) === true) return;
          printText(s, fp === 2 ? "#EE0000" : "");
        });
      } catch (ex) {
        console.error("api.listen:", ex);
      }
    });

    return {
      ...toRefs(state),
      pwd,
      scrollParent,
      inputElement,
      outputEmpty,
      autoscroll,
      showPlugins,
      onInputKey,
      clear,
      injectHook,
      load: api.load,
      save: api.save
    };
  }
};

function usePlugin(state) {
  const hooks = [] as any;
  const pluginStatus = ref([] as any);
  const hookInput = input => {
    // plugins are bash-only for now
    if (state.name != "bash") return false;
    const args = input.split(" ");
    for (const hook of hooks) {
      if (hook(args)) return true;
    }
    return false;
  };
  const injectHook = (hook, active) => {
    for (const one of hooks) {
      if (one === hook) return;
    }
    hooks.push(hook);
    pluginStatus.value.push(active);
  };
  return {
    hookInput,
    injectHook,
    pluginStatus
  };
}

function usePrintText(state, bufferLimit, autoscroll, scrollToEnd) {
  const pwd = ref("");
  api.pwd().then(value => (pwd.value = value));

  const inputFeedback = async input => {
    state.output += `<span style="color:gray; float:right">${_.escape(
      input
    )}</span>\n`;
    if (autoscroll.value === true) {
      scrollToEnd();
    }
    pwd.value = await api.pwd();
  };

  const printText = (s: string, color?: string) => {
    let output =
      state.output +
      (color
        ? `<span style="color:${color}">${_.escape(s)}</span>`
        : _.escape(s));

    // buffer limit
    if (output.length > bufferLimit * 1.2) {
      for (let i = output.length - bufferLimit - 1; i >= 0; i--) {
        if (output[i] == "\n") {
          output = output.substr(i + 1);
        }
      }
    }

    state.output = output;
    if (autoscroll.value === true) {
      scrollToEnd();
    }
  };
  return {
    inputFeedback,
    printText,
    pwd
  };
}

function useScrollUI(scrollParent) {
  const autoscroll = ref(true);
  let lastScroll: number | null = null;

  const scrollToEnd = () => {
    if (!scrollParent.value) {
      return;
    }
    if (lastScroll !== null) {
      clearTimeout(lastScroll);
    }
    lastScroll = setTimeout(
      () => (scrollParent.value.scrollTop = scrollParent.value.scrollHeight),
      50
    );
  };

  const scrollToStart = () => {
    if (!scrollParent.value) {
      return;
    }
    if (lastScroll !== null) {
      clearTimeout(lastScroll);
    }
    scrollParent.value.scrollTop = 0;
  };

  return {
    autoscroll,
    scrollToEnd,
    scrollToStart
  };
}

function useHistory(state) {
  let top = "";
  let index = -1;

  const STORAGE_KEY = "cli-history";
  const LIMIT = 10000;
  let history = JSON.parse(
    localStorage.getItem(STORAGE_KEY) || "[]"
  ) as string[];

  const back = (): string | null => {
    if (history.length == 0) return null;
    switch (index) {
      case -1:
        top = state.input;
        index = history.length - 1;
        break;
      case 0:
        break;
      default:
        index--;
        break;
    }
    if (index < 0 || index >= history.length) {
      return "";
    }
    return history[index];
  };

  const forward = (): string | null => {
    if (history.length == 0) return null;
    switch (index) {
      case -1:
        return top;
      case history.length - 1:
        index = -1;
        return top;
      default:
        index++;
        break;
    }
    if (index < 0 || index >= history.length) {
      return "";
    }
    return history[index];
  };

  const push = (cmd: string) => {
    index = -1;
    top = "";
    history.push(cmd);
    if (history.length > LIMIT) history = history.slice(history.length - LIMIT);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(history));
  };

  return {
    back,
    forward,
    push
  };
}

function useUserInput(
  state,
  SIGINT,
  SIGKILL,
  autoscroll,
  clear,
  printText,
  inputFeedback,
  back,
  forward,
  push,
  inputElement,
  hookInput
) {
  const onInputKey = async (e: KeyboardEvent) => {
    // Enter
    if (e.keyCode === 13) {
      if (e.shiftKey) return;

      e.preventDefault();
      const currentInput = state.input;
      state.input = "";

      inputFeedback(currentInput);
      push(currentInput);
      if (hookInput && hookInput(currentInput) === true) {
        return;
      }

      // alias for server: ll
      let serverInput = currentInput;
      const alias = { ll: "ls -l" };
      for (let name in alias) {
        const prefix = name + " ";
        if (currentInput == name || currentInput.startsWith(prefix)) {
          serverInput = alias[name] + currentInput.substr(name.length);
        }
      }

      try {
        await api.write(serverInput + "\n");
      } catch (ex) {
        console.error("api.write:", ex);
      }
      return;
    }

    if (e.ctrlKey) {
      switch (e.keyCode) {
        // Ctrl + 9 -> SIGKILL
        case 57:
          inputFeedback("[SIGKILL]");
          SIGKILL();
          break;
        // Ctrl + C -> SIGINT
        case 67:
          inputFeedback("[SIGINT]");
          SIGINT();
          break;
        // Ctrl + K -> clear
        case 75:
          clear();
          break;
        // Ctrl + S -> stop autoscroll
        case 83:
          if (autoscroll.value === true) {
            inputFeedback("[autoscroll off]");
            autoscroll.value = false;
          }
          break;
        // Ctrl + Q -> resume autoscroll
        case 81:
          if (autoscroll.value === false) {
            autoscroll.value = true;
            inputFeedback("[autoscroll on ]");
          }
          break;
      }
      return;
    }

    let input;
    switch (e.keyCode) {
      // up
      case 38:
        input = back();
        if (input !== null) state.input = input;
        inputElement.value.selectionStart = inputElement.value.selectionEnd =
          state.input.length;
        e.preventDefault();
        break;
      // down
      case 40:
        input = forward();
        if (input !== null) state.input = input;
        inputElement.value.selectionStart = inputElement.value.selectionEnd =
          state.input.length;
        e.preventDefault();
        break;
    }
  };
  return {
    onInputKey
  };
}

function useSendSignal() {
  return {
    SIGINT: () => api.kill(api.SIGINT),
    SIGKILL: () => api.kill(api.SIGKILL)
  };
}

function useTerminalCommand(state, scrollToStart) {
  const utf8decoder = new TextDecoder();
  const cmdclear = utf8decoder.decode(
    // darwin
    new Uint8Array([27, 91, 72, 27, 91, 50, 74])
  );
  const cmdclear2 = utf8decoder.decode(
    // ubuntu
    new Uint8Array([27, 91, 51, 74, 27, 91, 72, 27, 91, 50, 74])
  );
  const cmdclear3 = utf8decoder.decode(
    // tmux
    new Uint8Array([27, 91, 72, 27, 91, 74])
  );

  const handleCommand = (s: string): boolean => {
    if (s.length > 100) return false;
    switch (s) {
      case cmdclear:
      case cmdclear2:
      case cmdclear3:
        clear();
        return true;
    }
    return false;
  };

  const clear = () => {
    state.output = "";
    if (scrollToStart) scrollToStart();
  };
  return {
    handleCommand,
    clear
  };
}

function useMock(state) {
  if (!api.mock) return;
  state.input = "pwd";
  state.output = `total 152
-rw-r--r--   1 leo  staff   1076 Jan 30 00:48 LICENSE
-rw-r--r--   1 leo  staff     76 Jan 30 00:48 README.md
drwxr-xr-x   3 leo  staff     96 Jan 30 00:48 browser
drwxr-xr-x   3 leo  staff     96 Jan 30 00:48 chrome
drwxr-xr-x   3 leo  staff     96 Jan 30 00:48 cmd
-rw-r--r--   1 leo  staff    547 Feb  1 00:07 context.go
drwxr-xr-x  13 leo  staff    416 Feb  3 21:26 examples
-rw-r--r--   1 leo  staff    821 Jan 30 00:48 function.go
-rw-r--r--   1 leo  staff    105 Feb  3 21:22 go.mod
-rw-r--r--   1 leo  staff    677 Feb  3 21:22 go.sum
drwxr-xr-x   3 leo  staff     96 Jan 30 00:48 homedir
-rw-r--r--   1 leo  staff   5105 Feb  3 18:43 jsclient.go
drwxr-xr-x   3 leo  staff     96 Jan 30 00:48 one
-rw-r--r--   1 leo  staff   5724 Feb  1 00:14 page.go
-rw-r--r--   1 leo  staff  12625 Feb  3 21:24 script.go
-rwxr-xr-x   1 leo  staff     44 Feb  3 19:23 sync
-rw-r--r--   1 leo  staff   1015 Jan 30 00:48 value.go
-rw-r--r--   1 leo  staff   7050 Feb  3 21:18 vue.go
drwxr-xr-x   5 leo  staff    160 Feb  1 01:33 vue2go
-rw-r--r--   1 leo  staff    226 Jan 30 00:48 window.go
`;
}
</script>

<style>
html,
body {
  margin: 0px;
  height: 100%;
}
</style>

<style scoped>
.root {
  width: 100%;
  height: 100%;
  font-family: "monaco", "monospace";
  font-size: 14px;
  display: flex;
}

.plugin {
  flex: 0 0 auto;
  width: 0;
  height: 100%;
  z-index: 1;
  background: white;
  visibility: hidden;
}

.plugin.show {
  width: 50%;
  visibility: visible;
}

.chat {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  height: 100%;
  color: #cccccc;
}

.out {
  flex: 1 1 auto;
  background: black;
  overflow-y: scroll;
  position: relative;
}

.out pre {
  margin: 0px;
  padding: 5px;
  line-height: 1.5em;
  font-family: inherit;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.out .title {
  position: absolute;
  top: calc(50% - 50px);
  width: 100%;
  height: 100px;
  font-size: 80px;
  color: red;
  opacity: 1;
  text-align: center;
  pointer-events: none;
  visibility: hidden;
}

.head {
  flex: 0 0 auto;
  padding: 0 5px 5px 5px;
  height: 20px;
  background: black;
  position: relative;
  color: gray;
}

.out .title.show {
  visibility: visible;
}

.in {
  background: black;
  flex: 0 0 auto;
  height: 200px;
}

.in textarea {
  background: inherit;
  color: inherit;
  width: calc(100% - 5px);
  height: calc(100% - 2 * 5px - 2px);
  padding: 5px;
  border-width: 2px 0 0 0;
  border-style: solid;
  font-size: inherit;
  font-family: inherit;
  opacity: 0.618;
}

.in textarea:focus {
  opacity: 1;
}
</style>

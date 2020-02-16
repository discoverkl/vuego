<template>
  <div class="root">
    <!-- <textarea
      class="editor"
      v-model="text"
      @keydown="onInputKey"
      ref="inputElement"
    /> -->
    <div class="editor" ref="inputElement" @keydown="onInputKey"></div>
    <!-- <div class="status" ref="statusElement">some text</div> -->
    <div class="error">
      {{ err }}
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "underscore";
import {
  ref,
  computed,
  watch,
  reactive,
  toRefs,
  onMounted
} from "@vue/composition-api";
import * as monaco from "monaco-editor";

export default {
  props: ["load", "save"],
  setup(props, { emit }) {
    const state = reactive({
      path: "",
      text: "",
      err: ""
    });
    const inputElement = ref(null as HTMLElement | null);
    const lastLoadedText = ref("");
    const active = computed(
      () => state.path !== "" || state.err !== "" || false
    );

    const { hook } = useHook(props, state, active, lastLoadedText);

    let editor: monaco.editor.IStandaloneCodeEditor | null = null;

    // open file
    watch(async () => {
      try {
        if (state.path === "") {
          return;
        }
        state.text = await props.load(state.path);
      } catch (ex) {
        state.text = "";
        state.err = ex.toString();
      }
      lastLoadedText.value = state.text;

      // update editor
      if (editor !== null) {
        const lang = filename2language(state.path);
        editor.setModel(monaco.editor.createModel(state.text, lang));
        editor.getModel()!.onDidChangeContent(e => {
          state.text = editor!.getValue();
        });
        editor.layout();
        editor.focus();
      }
    });

    // enable plugin
    onMounted(() => {
      emit("provideHook", hook, active);

      editor = monaco.editor.create(inputElement.value as any, {
        theme: "vs-dark",
        scrollBeyondLastLine: false
        // automaticLayout: true
      });

      useBodyResize(editor, active);

      // bind editor text to state.text
      // (window as any).e = editor;
      // (window as any).m = monaco;
    });

    const languages = monaco.languages.getLanguages();

    const filename2language = (name: string) => {
      name = name.substr(name.lastIndexOf("/") + 1);
      const index = name.lastIndexOf(".");
      if (index === -1) {
        return "";
      }
      const ext = name.substr(index).toLowerCase();
      for (const lang of languages) {
        if (lang.extensions && lang.extensions.indexOf(ext) !== -1) {
          return lang.id;
        }
      }
    };

    const onInputKey = (e: KeyboardEvent) => {
      // Esc
      if (e.keyCode == 27) {
        if (e.ctrlKey || e.metaKey) return;
        e.preventDefault();
        emit("escape");
      }
    };

    return {
      ...toRefs(state),
      inputElement,
      onInputKey
    };
  }
};

function useBodyResize(editor, active) {
  const lazyLayout = _.debounce(() => {
    // console.log("editor resize");
    editor.layout();
  }, 500);
  window.document.body.onresize = () => {
    // console.log("body resize");
    if (!active.value) return;
    lazyLayout();
  };
}

function useHook(props, state, active, lastLoadedText) {
  const hook = args => {
    let save = false;
    let quit = false;
    let force = false;
    let open = false;

    switch (args[0]) {
      case "vi":
      case "vim":
        open = true;
        break;
      case ":q":
        quit = true;
        break;
      case ":w":
        save = true;
        break;
      case ":wq":
        quit = save = true;
        break;
      case ":q!":
        quit = force = true;
        break;
      default:
        return false;
    }

    const onCommand = async () => {
      if (save) {
        try {
          await props.save(state.path, state.text);
        } catch (ex) {
          state.err = ex.toString();
          return;
        }
      }
      if (quit) {
        if (!save && !force && lastLoadedText.value != state.text) {
          state.err = "No write since last change (add ! to override)";
          return;
        }
        state.path = state.err = "";
      }
    };

    // process ':' commands
    if (!open) {
      if (active.value === false) return false;
      onCommand();
      return true;
    }

    // process vi/vim command
    state.path = "";
    state.err = "";

    if (args.length != 2) {
      state.err = `vim: invalid arguments: ${args.slice(1)}`;
      return true;
    }

    state.path = args[1];

    return true;
  };
  return {
    hook
  };
}
</script>

<style scoped>
.root {
  width: 100%;
  height: 100%;
  position: relative;
  /* display: flex;
  flex-direction: column; */
}

.editor {
  width: 100%;
  /* height: calc(100% - 2 * 5px - 25px); */
  height: calc(100% - 2 * 5px);
  /* font-size: inherit;
  font-family: inherit; */
  padding: 5px;
  /* background: red; */
  background: #1e1e1e;
}
/* 
.status {
  height: 25px;
  z-index: 1;
  background: white;
  padding: 4px;
} */

.error {
  position: absolute;
  right: 0;
  bottom: 0;
  margin: 20px;
  color: red;
  pointer-events: none;
  /* font-size: 20px; */
  max-width: calc(100% - 2 * 20px);
  word-wrap: break-word;
}
</style>

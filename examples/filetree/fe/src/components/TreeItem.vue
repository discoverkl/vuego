<template>
  <li>
    <div :class="{ bold: isFolder }" @click="toggle">
      {{ baseName }}
      <span v-if="isFolder">[{{ isOpen ? "-" : "+" }}]</span>
    </div>
    <ul v-show="isOpen" v-if="isFolder">
      <tree-item
        class="item"
        v-for="(child, index) in item.children"
        :key="version.toString() + '-' + index"
        :item="child"
      ></tree-item>
    </ul>
  </li>
</template>

<script lang="ts">
import Vue from "vue";
import { ref, reactive, toRefs, computed } from "@vue/composition-api";

interface Folder {
  name: string;
  children: Folder[];
  isFolder: boolean;
}

// openFolder is implemented in Go
declare function openFolder(path: string): Folder;

interface Props {
  item: Folder;
}

export default {
  name: "tree-item",
  props: {
    item: Object
  },
  setup(props: Props) {
    const state = reactive({
      isOpen: false,
      isFolder: computed(() => props.item.isFolder),
      baseName: computed(() =>
        props.item.name === "/"
          ? "My Computer"
          : props.item.name.substr(props.item.name.lastIndexOf("/") + 1)
      ),
      version: 0
    });

    async function toggle() {
      if (state.isFolder) {
        if (!state.isOpen) {
          try {
            state.version++
            Vue.set(
              props.item,
              "children",
              (await openFolder(props.item.name)).children
            );
          } catch (ex) {
            Vue.set(props.item, "children", null);
            console.log("openFolder failed:", ex);
          }
        }

        state.isOpen = !state.isOpen;
      }
    }

    return {
      ...toRefs(state),
      toggle
    };
  }
};
</script>

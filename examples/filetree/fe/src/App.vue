<template>
  <div>
    <p>(Your file system via Go API.)</p>

    <ul>
      <tree-item class="item" :item="treeData"></tree-item>
    </ul>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import {
  ref,
  reactive,
  toRefs,
  computed,
  onMounted
} from "@vue/composition-api";
import TreeItem from "./components/TreeItem";

export default {
  components: {
    "tree-item": TreeItem
  },
  setup() {
    const state = reactive({
      treeData: {
        name: "/",
        isFolder: true
      }
    });

    return {
      ...toRefs(state)
    };
  }
};
</script>

<style>
body {
  font-family: Menlo, Consolas, monospace;
  color: #444;
}
.item {
  cursor: pointer;
}
.bold {
  font-weight: bold;
}
ul {
  padding-left: 1em;
  line-height: 1.5em;
}
</style>

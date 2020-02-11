module.exports = {
  extends: ["@vue/typescript", "plugin:vue/essential", "@vue/prettier"],

  parserOptions: {
    parser: "@typescript-eslint/parser"
  },

  root: true,

  env: {
    node: true
  },

  rules: {
    "no-unused-vars": "off",
    "no-console": "off",
    "no-debugger": process.env.NODE_ENV === "production" ? "error" : "off"
  }
};

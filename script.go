package vuego

import (
	"fmt"
	"encoding/json"
	"strings"
)

type jsOption struct {
	Dev bool `json:"dev"`
	ReadyFuncName string `json:"readyFuncName"`
	Prefix string `json:"prefix"`
	Search string `json:"search"`
	Bindings []string `json:"bindings"`
}

func init() {
	// script = mapScript(script, "let dev = true", "let dev = false")
	// script = mapScript(script, "Vuego()", fmt.Sprintf("%s()", ReadyFuncName))
}

func injectOptions(op *jsOption) string {
	if op == nil {
		op = &jsOption{}
	}
	op.Dev = false	
	op.ReadyFuncName = ReadyFuncName
	raw, _ := json.MarshalIndent(op, "    ", "    ")
	text := string(raw)
	return mapScript(script, "options = null", fmt.Sprintf("options = %s", text))
}

func mapScript(in, old, new string) string {
	if old == new {
		return in
	}
	index := strings.Index(in, old)
	if index == -1 {
		panic(fmt.Sprintf("mspScript error, old string not found: %s", old))
	}
	ret := strings.Replace(in, old, new, 1)
	index = strings.Index(ret, old)
	if index != -1 {
		panic(fmt.Sprintf("mspScript error, old string appears many times: %s", old))
	}
	return ret
}

var script = `/******/ (function(modules) { // webpackBootstrap
	/******/ 	// install a JSONP callback for chunk loading
	/******/ 	function webpackJsonpCallback(data) {
	/******/ 		var chunkIds = data[0];
	/******/ 		var moreModules = data[1];
	/******/ 		var executeModules = data[2];
	/******/
	/******/ 		// add "moreModules" to the modules object,
	/******/ 		// then flag all "chunkIds" as loaded and fire callback
	/******/ 		var moduleId, chunkId, i = 0, resolves = [];
	/******/ 		for(;i < chunkIds.length; i++) {
	/******/ 			chunkId = chunkIds[i];
	/******/ 			if(Object.prototype.hasOwnProperty.call(installedChunks, chunkId) && installedChunks[chunkId]) {
	/******/ 				resolves.push(installedChunks[chunkId][0]);
	/******/ 			}
	/******/ 			installedChunks[chunkId] = 0;
	/******/ 		}
	/******/ 		for(moduleId in moreModules) {
	/******/ 			if(Object.prototype.hasOwnProperty.call(moreModules, moduleId)) {
	/******/ 				modules[moduleId] = moreModules[moduleId];
	/******/ 			}
	/******/ 		}
	/******/ 		if(parentJsonpFunction) parentJsonpFunction(data);
	/******/
	/******/ 		while(resolves.length) {
	/******/ 			resolves.shift()();
	/******/ 		}
	/******/
	/******/ 		// add entry modules from loaded chunk to deferred list
	/******/ 		deferredModules.push.apply(deferredModules, executeModules || []);
	/******/
	/******/ 		// run deferred modules when all chunks ready
	/******/ 		return checkDeferredModules();
	/******/ 	};
	/******/ 	function checkDeferredModules() {
	/******/ 		var result;
	/******/ 		for(var i = 0; i < deferredModules.length; i++) {
	/******/ 			var deferredModule = deferredModules[i];
	/******/ 			var fulfilled = true;
	/******/ 			for(var j = 1; j < deferredModule.length; j++) {
	/******/ 				var depId = deferredModule[j];
	/******/ 				if(installedChunks[depId] !== 0) fulfilled = false;
	/******/ 			}
	/******/ 			if(fulfilled) {
	/******/ 				deferredModules.splice(i--, 1);
	/******/ 				result = __webpack_require__(__webpack_require__.s = deferredModule[0]);
	/******/ 			}
	/******/ 		}
	/******/
	/******/ 		return result;
	/******/ 	}
	/******/
	/******/ 	// The module cache
	/******/ 	var installedModules = {};
	/******/
	/******/ 	// object to store loaded and loading chunks
	/******/ 	// undefined = chunk not loaded, null = chunk preloaded/prefetched
	/******/ 	// Promise = chunk loading, 0 = chunk loaded
	/******/ 	var installedChunks = {
	/******/ 		"app": 0
	/******/ 	};
	/******/
	/******/ 	var deferredModules = [];
	/******/
	/******/ 	// The require function
	/******/ 	function __webpack_require__(moduleId) {
	/******/
	/******/ 		// Check if module is in cache
	/******/ 		if(installedModules[moduleId]) {
	/******/ 			return installedModules[moduleId].exports;
	/******/ 		}
	/******/ 		// Create a new module (and put it into the cache)
	/******/ 		var module = installedModules[moduleId] = {
	/******/ 			i: moduleId,
	/******/ 			l: false,
	/******/ 			exports: {}
	/******/ 		};
	/******/
	/******/ 		// Execute the module function
	/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
	/******/
	/******/ 		// Flag the module as loaded
	/******/ 		module.l = true;
	/******/
	/******/ 		// Return the exports of the module
	/******/ 		return module.exports;
	/******/ 	}
	/******/
	/******/
	/******/ 	// expose the modules object (__webpack_modules__)
	/******/ 	__webpack_require__.m = modules;
	/******/
	/******/ 	// expose the module cache
	/******/ 	__webpack_require__.c = installedModules;
	/******/
	/******/ 	// define getter function for harmony exports
	/******/ 	__webpack_require__.d = function(exports, name, getter) {
	/******/ 		if(!__webpack_require__.o(exports, name)) {
	/******/ 			Object.defineProperty(exports, name, { enumerable: true, get: getter });
	/******/ 		}
	/******/ 	};
	/******/
	/******/ 	// define __esModule on exports
	/******/ 	__webpack_require__.r = function(exports) {
	/******/ 		if(typeof Symbol !== 'undefined' && Symbol.toStringTag) {
	/******/ 			Object.defineProperty(exports, Symbol.toStringTag, { value: 'Module' });
	/******/ 		}
	/******/ 		Object.defineProperty(exports, '__esModule', { value: true });
	/******/ 	};
	/******/
	/******/ 	// create a fake namespace object
	/******/ 	// mode & 1: value is a module id, require it
	/******/ 	// mode & 2: merge all properties of value into the ns
	/******/ 	// mode & 4: return value when already ns object
	/******/ 	// mode & 8|1: behave like require
	/******/ 	__webpack_require__.t = function(value, mode) {
	/******/ 		if(mode & 1) value = __webpack_require__(value);
	/******/ 		if(mode & 8) return value;
	/******/ 		if((mode & 4) && typeof value === 'object' && value && value.__esModule) return value;
	/******/ 		var ns = Object.create(null);
	/******/ 		__webpack_require__.r(ns);
	/******/ 		Object.defineProperty(ns, 'default', { enumerable: true, value: value });
	/******/ 		if(mode & 2 && typeof value != 'string') for(var key in value) __webpack_require__.d(ns, key, function(key) { return value[key]; }.bind(null, key));
	/******/ 		return ns;
	/******/ 	};
	/******/
	/******/ 	// getDefaultExport function for compatibility with non-harmony modules
	/******/ 	__webpack_require__.n = function(module) {
	/******/ 		var getter = module && module.__esModule ?
	/******/ 			function getDefault() { return module['default']; } :
	/******/ 			function getModuleExports() { return module; };
	/******/ 		__webpack_require__.d(getter, 'a', getter);
	/******/ 		return getter;
	/******/ 	};
	/******/
	/******/ 	// Object.prototype.hasOwnProperty.call
	/******/ 	__webpack_require__.o = function(object, property) { return Object.prototype.hasOwnProperty.call(object, property); };
	/******/
	/******/ 	// __webpack_public_path__
	/******/ 	__webpack_require__.p = "/";
	/******/
	/******/ 	var jsonpArray = window["webpackJsonp"] = window["webpackJsonp"] || [];
	/******/ 	var oldJsonpFunction = jsonpArray.push.bind(jsonpArray);
	/******/ 	jsonpArray.push = webpackJsonpCallback;
	/******/ 	jsonpArray = jsonpArray.slice();
	/******/ 	for(var i = 0; i < jsonpArray.length; i++) webpackJsonpCallback(jsonpArray[i]);
	/******/ 	var parentJsonpFunction = oldJsonpFunction;
	/******/
	/******/
	/******/ 	// add entry module to deferred list
	/******/ 	deferredModules.push([0,"chunk-vendors"]);
	/******/ 	// run deferred modules when ready
	/******/ 	return checkDeferredModules();
	/******/ })
	/************************************************************************/
	/******/ ({
	
	/***/ "./node_modules/cache-loader/dist/cjs.js?!./node_modules/ts-loader/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=script&lang=ts&":
	/*!*****************************************************************************************************************************************************************************************************************************************!*\
	  !*** ./node_modules/cache-loader/dist/cjs.js??ref--13-0!./node_modules/ts-loader??ref--13-1!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/App.vue?vue&type=script&lang=ts& ***!
	  \*****************************************************************************************************************************************************************************************************************************************/
	/*! exports provided: default */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _vue_composition_api__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! @vue/composition-api */ \"./node_modules/@vue/composition-api/dist/vue-composition-api.module.js\");\n/* harmony import */ var _components_TreeItem_vue__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ./components/TreeItem.vue */ \"./src/components/TreeItem.vue\");\n\n\n/* harmony default export */ __webpack_exports__[\"default\"] = ({\n    components: {\n        \"tree-item\": _components_TreeItem_vue__WEBPACK_IMPORTED_MODULE_1__[\"default\"]\n    },\n    setup() {\n        const state = Object(_vue_composition_api__WEBPACK_IMPORTED_MODULE_0__[\"reactive\"])({\n            treeData: {\n                name: \"/\",\n                isFolder: true\n            }\n        });\n        return {\n            ...Object(_vue_composition_api__WEBPACK_IMPORTED_MODULE_0__[\"toRefs\"])(state)\n        };\n    }\n});\n\n\n//# sourceURL=webpack:///./src/App.vue?./node_modules/cache-loader/dist/cjs.js??ref--13-0!./node_modules/ts-loader??ref--13-1!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options");
	
	/***/ }),
	
	/***/ "./node_modules/cache-loader/dist/cjs.js?!./node_modules/ts-loader/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/components/TreeItem.vue?vue&type=script&lang=ts&":
	/*!*********************************************************************************************************************************************************************************************************************************************************!*\
	  !*** ./node_modules/cache-loader/dist/cjs.js??ref--13-0!./node_modules/ts-loader??ref--13-1!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/components/TreeItem.vue?vue&type=script&lang=ts& ***!
	  \*********************************************************************************************************************************************************************************************************************************************************/
	/*! exports provided: default */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var vue__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! vue */ \"./node_modules/vue/dist/vue.runtime.esm.js\");\n/* harmony import */ var _vue_composition_api__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @vue/composition-api */ \"./node_modules/@vue/composition-api/dist/vue-composition-api.module.js\");\n/* harmony import */ var vue2go__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! vue2go */ \"./node_modules/vue2go/vue2go.js\");\n/* harmony import */ var vue2go__WEBPACK_IMPORTED_MODULE_2___default = /*#__PURE__*/__webpack_require__.n(vue2go__WEBPACK_IMPORTED_MODULE_2__);\n\n\n\nconst api = Object(vue2go__WEBPACK_IMPORTED_MODULE_2__[\"getapi\"])();\n/* harmony default export */ __webpack_exports__[\"default\"] = ({\n    name: \"tree-item\",\n    props: {\n        item: Object\n    },\n    setup(props) {\n        const state = Object(_vue_composition_api__WEBPACK_IMPORTED_MODULE_1__[\"reactive\"])({\n            isOpen: false,\n            isFolder: Object(_vue_composition_api__WEBPACK_IMPORTED_MODULE_1__[\"computed\"])(() => props.item.isFolder),\n            baseName: Object(_vue_composition_api__WEBPACK_IMPORTED_MODULE_1__[\"computed\"])(() => props.item.name === \"/\"\n                ? \"My Computer\"\n                : props.item.name.substr(props.item.name.lastIndexOf(\"/\") + 1)),\n            version: 0\n        });\n        async function toggle() {\n            if (state.isFolder) {\n                if (!state.isOpen) {\n                    try {\n                        state.version++;\n                        vue__WEBPACK_IMPORTED_MODULE_0__[\"default\"].set(props.item, \"children\", (await api.openFolder(props.item.name)).children);\n                    }\n                    catch (ex) {\n                        vue__WEBPACK_IMPORTED_MODULE_0__[\"default\"].set(props.item, \"children\", null);\n                        console.error(\"openFolder failed:\", ex);\n                    }\n                }\n                state.isOpen = !state.isOpen;\n            }\n        }\n        return {\n            ...Object(_vue_composition_api__WEBPACK_IMPORTED_MODULE_1__[\"toRefs\"])(state),\n            toggle\n        };\n    }\n});\n\n\n//# sourceURL=webpack:///./src/components/TreeItem.vue?./node_modules/cache-loader/dist/cjs.js??ref--13-0!./node_modules/ts-loader??ref--13-1!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options");
	
	/***/ }),
	
	/***/ "./node_modules/cache-loader/dist/cjs.js?{\"cacheDirectory\":\"node_modules/.cache/vue-loader\",\"cacheIdentifier\":\"3c385f58-vue-loader-template\"}!./node_modules/vue-loader/lib/loaders/templateLoader.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=template&id=7ba5bd90&":
	/*!*********************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
	  !*** ./node_modules/cache-loader/dist/cjs.js?{"cacheDirectory":"node_modules/.cache/vue-loader","cacheIdentifier":"3c385f58-vue-loader-template"}!./node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/App.vue?vue&type=template&id=7ba5bd90& ***!
	  \*********************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************/
	/*! exports provided: render, staticRenderFns */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"render\", function() { return render; });\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"staticRenderFns\", function() { return staticRenderFns; });\nvar render = function() {\n  var _vm = this\n  var _h = _vm.$createElement\n  var _c = _vm._self._c || _h\n  return _c(\"div\", [\n    _c(\"p\", [_vm._v(\"(Your file system via Go API.)\")]),\n    _c(\n      \"ul\",\n      [_c(\"tree-item\", { staticClass: \"item\", attrs: { item: _vm.treeData } })],\n      1\n    )\n  ])\n}\nvar staticRenderFns = []\nrender._withStripped = true\n\n\n\n//# sourceURL=webpack:///./src/App.vue?./node_modules/cache-loader/dist/cjs.js?%7B%22cacheDirectory%22:%22node_modules/.cache/vue-loader%22,%22cacheIdentifier%22:%223c385f58-vue-loader-template%22%7D!./node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options");
	
	/***/ }),
	
	/***/ "./node_modules/cache-loader/dist/cjs.js?{\"cacheDirectory\":\"node_modules/.cache/vue-loader\",\"cacheIdentifier\":\"3c385f58-vue-loader-template\"}!./node_modules/vue-loader/lib/loaders/templateLoader.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/components/TreeItem.vue?vue&type=template&id=1c2a3381&":
	/*!*************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
	  !*** ./node_modules/cache-loader/dist/cjs.js?{"cacheDirectory":"node_modules/.cache/vue-loader","cacheIdentifier":"3c385f58-vue-loader-template"}!./node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/components/TreeItem.vue?vue&type=template&id=1c2a3381& ***!
	  \*************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************/
	/*! exports provided: render, staticRenderFns */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"render\", function() { return render; });\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"staticRenderFns\", function() { return staticRenderFns; });\nvar render = function() {\n  var _vm = this\n  var _h = _vm.$createElement\n  var _c = _vm._self._c || _h\n  return _c(\"li\", [\n    _c(\"div\", { class: { bold: _vm.isFolder }, on: { click: _vm.toggle } }, [\n      _vm._v(\" \" + _vm._s(_vm.baseName) + \" \"),\n      _vm.isFolder\n        ? _c(\"span\", [_vm._v(\"[\" + _vm._s(_vm.isOpen ? \"-\" : \"+\") + \"]\")])\n        : _vm._e()\n    ]),\n    _vm.isFolder\n      ? _c(\n          \"ul\",\n          {\n            directives: [\n              {\n                name: \"show\",\n                rawName: \"v-show\",\n                value: _vm.isOpen,\n                expression: \"isOpen\"\n              }\n            ]\n          },\n          _vm._l(_vm.item.children, function(child, index) {\n            return _c(\"tree-item\", {\n              key: _vm.version.toString() + \"-\" + index,\n              staticClass: \"item\",\n              attrs: { item: child }\n            })\n          }),\n          1\n        )\n      : _vm._e()\n  ])\n}\nvar staticRenderFns = []\nrender._withStripped = true\n\n\n\n//# sourceURL=webpack:///./src/components/TreeItem.vue?./node_modules/cache-loader/dist/cjs.js?%7B%22cacheDirectory%22:%22node_modules/.cache/vue-loader%22,%22cacheIdentifier%22:%223c385f58-vue-loader-template%22%7D!./node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options");
	
	/***/ }),
	
	/***/ "./node_modules/css-loader/dist/cjs.js?!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=style&index=0&lang=css&":
	/*!*******************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
	  !*** ./node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src??ref--6-oneOf-1-2!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/App.vue?vue&type=style&index=0&lang=css& ***!
	  \*******************************************************************************************************************************************************************************************************************************************************************************************************************************/
	/*! no static exports found */
	/***/ (function(module, exports, __webpack_require__) {
	
	eval("// Imports\nvar ___CSS_LOADER_API_IMPORT___ = __webpack_require__(/*! ../node_modules/css-loader/dist/runtime/api.js */ \"./node_modules/css-loader/dist/runtime/api.js\");\nexports = ___CSS_LOADER_API_IMPORT___(false);\n// Module\nexports.push([module.i, \"\\nbody {\\n  font-family: Menlo, Consolas, monospace;\\n  color: #444;\\n}\\n.item {\\n  cursor: pointer;\\n}\\n.bold {\\n  font-weight: bold;\\n}\\nul {\\n  padding-left: 1em;\\n  line-height: 1.5em;\\n}\\n\", \"\"]);\n// Exports\nmodule.exports = exports;\n\n\n//# sourceURL=webpack:///./src/App.vue?./node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src??ref--6-oneOf-1-2!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options");
	
	/***/ }),
	
	/***/ "./node_modules/vue-style-loader/index.js?!./node_modules/css-loader/dist/cjs.js?!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=style&index=0&lang=css&":
	/*!*********************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
	  !*** ./node_modules/vue-style-loader??ref--6-oneOf-1-0!./node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src??ref--6-oneOf-1-2!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options!./src/App.vue?vue&type=style&index=0&lang=css& ***!
	  \*********************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************/
	/*! no static exports found */
	/***/ (function(module, exports, __webpack_require__) {
	
	eval("// style-loader: Adds some css to the DOM by adding a <style> tag\n\n// load the styles\nvar content = __webpack_require__(/*! !../node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!../node_modules/vue-loader/lib/loaders/stylePostLoader.js!../node_modules/postcss-loader/src??ref--6-oneOf-1-2!../node_modules/cache-loader/dist/cjs.js??ref--0-0!../node_modules/vue-loader/lib??vue-loader-options!./App.vue?vue&type=style&index=0&lang=css& */ \"./node_modules/css-loader/dist/cjs.js?!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=style&index=0&lang=css&\");\nif(typeof content === 'string') content = [[module.i, content, '']];\nif(content.locals) module.exports = content.locals;\n// add the styles to the DOM\nvar add = __webpack_require__(/*! ../node_modules/vue-style-loader/lib/addStylesClient.js */ \"./node_modules/vue-style-loader/lib/addStylesClient.js\").default\nvar update = add(\"fa1ef42a\", content, false, {\"sourceMap\":false,\"shadowMode\":false});\n// Hot Module Replacement\nif(false) {}\n\n//# sourceURL=webpack:///./src/App.vue?./node_modules/vue-style-loader??ref--6-oneOf-1-0!./node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src??ref--6-oneOf-1-2!./node_modules/cache-loader/dist/cjs.js??ref--0-0!./node_modules/vue-loader/lib??vue-loader-options");
	
	/***/ }),
	
	/***/ "./src/App.vue":
	/*!*********************!*\
	  !*** ./src/App.vue ***!
	  \*********************/
	/*! exports provided: default */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _App_vue_vue_type_template_id_7ba5bd90___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./App.vue?vue&type=template&id=7ba5bd90& */ \"./src/App.vue?vue&type=template&id=7ba5bd90&\");\n/* harmony import */ var _App_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ./App.vue?vue&type=script&lang=ts& */ \"./src/App.vue?vue&type=script&lang=ts&\");\n/* empty/unused harmony star reexport *//* harmony import */ var _App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ./App.vue?vue&type=style&index=0&lang=css& */ \"./src/App.vue?vue&type=style&index=0&lang=css&\");\n/* harmony import */ var _node_modules_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../node_modules/vue-loader/lib/runtime/componentNormalizer.js */ \"./node_modules/vue-loader/lib/runtime/componentNormalizer.js\");\n\n\n\n\n\n\n/* normalize component */\n\nvar component = Object(_node_modules_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_3__[\"default\"])(\n  _App_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_1__[\"default\"],\n  _App_vue_vue_type_template_id_7ba5bd90___WEBPACK_IMPORTED_MODULE_0__[\"render\"],\n  _App_vue_vue_type_template_id_7ba5bd90___WEBPACK_IMPORTED_MODULE_0__[\"staticRenderFns\"],\n  false,\n  null,\n  null,\n  null\n  \n)\n\n/* hot reload */\nif (false) { var api; }\ncomponent.options.__file = \"src/App.vue\"\n/* harmony default export */ __webpack_exports__[\"default\"] = (component.exports);\n\n//# sourceURL=webpack:///./src/App.vue?");
	
	/***/ }),
	
	/***/ "./src/App.vue?vue&type=script&lang=ts&":
	/*!**********************************************!*\
	  !*** ./src/App.vue?vue&type=script&lang=ts& ***!
	  \**********************************************/
	/*! exports provided: default */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _node_modules_cache_loader_dist_cjs_js_ref_13_0_node_modules_ts_loader_index_js_ref_13_1_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../node_modules/cache-loader/dist/cjs.js??ref--13-0!../node_modules/ts-loader??ref--13-1!../node_modules/cache-loader/dist/cjs.js??ref--0-0!../node_modules/vue-loader/lib??vue-loader-options!./App.vue?vue&type=script&lang=ts& */ \"./node_modules/cache-loader/dist/cjs.js?!./node_modules/ts-loader/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=script&lang=ts&\");\n/* empty/unused harmony star reexport */ /* harmony default export */ __webpack_exports__[\"default\"] = (_node_modules_cache_loader_dist_cjs_js_ref_13_0_node_modules_ts_loader_index_js_ref_13_1_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_0__[\"default\"]); \n\n//# sourceURL=webpack:///./src/App.vue?");
	
	/***/ }),
	
	/***/ "./src/App.vue?vue&type=style&index=0&lang=css&":
	/*!******************************************************!*\
	  !*** ./src/App.vue?vue&type=style&index=0&lang=css& ***!
	  \******************************************************/
	/*! no static exports found */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _node_modules_vue_style_loader_index_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../node_modules/vue-style-loader??ref--6-oneOf-1-0!../node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!../node_modules/vue-loader/lib/loaders/stylePostLoader.js!../node_modules/postcss-loader/src??ref--6-oneOf-1-2!../node_modules/cache-loader/dist/cjs.js??ref--0-0!../node_modules/vue-loader/lib??vue-loader-options!./App.vue?vue&type=style&index=0&lang=css& */ \"./node_modules/vue-style-loader/index.js?!./node_modules/css-loader/dist/cjs.js?!./node_modules/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/postcss-loader/src/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=style&index=0&lang=css&\");\n/* harmony import */ var _node_modules_vue_style_loader_index_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0___default = /*#__PURE__*/__webpack_require__.n(_node_modules_vue_style_loader_index_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__);\n/* harmony reexport (unknown) */ for(var __WEBPACK_IMPORT_KEY__ in _node_modules_vue_style_loader_index_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__) if(__WEBPACK_IMPORT_KEY__ !== 'default') (function(key) { __webpack_require__.d(__webpack_exports__, key, function() { return _node_modules_vue_style_loader_index_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__[key]; }) }(__WEBPACK_IMPORT_KEY__));\n /* harmony default export */ __webpack_exports__[\"default\"] = (_node_modules_vue_style_loader_index_js_ref_6_oneOf_1_0_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_node_modules_vue_loader_lib_loaders_stylePostLoader_js_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_2_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0___default.a); \n\n//# sourceURL=webpack:///./src/App.vue?");
	
	/***/ }),
	
	/***/ "./src/App.vue?vue&type=template&id=7ba5bd90&":
	/*!****************************************************!*\
	  !*** ./src/App.vue?vue&type=template&id=7ba5bd90& ***!
	  \****************************************************/
	/*! exports provided: render, staticRenderFns */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _node_modules_cache_loader_dist_cjs_js_cacheDirectory_node_modules_cache_vue_loader_cacheIdentifier_3c385f58_vue_loader_template_node_modules_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_template_id_7ba5bd90___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../node_modules/cache-loader/dist/cjs.js?{\"cacheDirectory\":\"node_modules/.cache/vue-loader\",\"cacheIdentifier\":\"3c385f58-vue-loader-template\"}!../node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!../node_modules/cache-loader/dist/cjs.js??ref--0-0!../node_modules/vue-loader/lib??vue-loader-options!./App.vue?vue&type=template&id=7ba5bd90& */ \"./node_modules/cache-loader/dist/cjs.js?{\\\"cacheDirectory\\\":\\\"node_modules/.cache/vue-loader\\\",\\\"cacheIdentifier\\\":\\\"3c385f58-vue-loader-template\\\"}!./node_modules/vue-loader/lib/loaders/templateLoader.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/App.vue?vue&type=template&id=7ba5bd90&\");\n/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, \"render\", function() { return _node_modules_cache_loader_dist_cjs_js_cacheDirectory_node_modules_cache_vue_loader_cacheIdentifier_3c385f58_vue_loader_template_node_modules_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_template_id_7ba5bd90___WEBPACK_IMPORTED_MODULE_0__[\"render\"]; });\n\n/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, \"staticRenderFns\", function() { return _node_modules_cache_loader_dist_cjs_js_cacheDirectory_node_modules_cache_vue_loader_cacheIdentifier_3c385f58_vue_loader_template_node_modules_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_App_vue_vue_type_template_id_7ba5bd90___WEBPACK_IMPORTED_MODULE_0__[\"staticRenderFns\"]; });\n\n\n\n//# sourceURL=webpack:///./src/App.vue?");
	
	/***/ }),
	
	/***/ "./src/components/TreeItem.vue":
	/*!*************************************!*\
	  !*** ./src/components/TreeItem.vue ***!
	  \*************************************/
	/*! exports provided: default */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _TreeItem_vue_vue_type_template_id_1c2a3381___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./TreeItem.vue?vue&type=template&id=1c2a3381& */ \"./src/components/TreeItem.vue?vue&type=template&id=1c2a3381&\");\n/* harmony import */ var _TreeItem_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ./TreeItem.vue?vue&type=script&lang=ts& */ \"./src/components/TreeItem.vue?vue&type=script&lang=ts&\");\n/* empty/unused harmony star reexport *//* harmony import */ var _node_modules_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ../../node_modules/vue-loader/lib/runtime/componentNormalizer.js */ \"./node_modules/vue-loader/lib/runtime/componentNormalizer.js\");\n\n\n\n\n\n/* normalize component */\n\nvar component = Object(_node_modules_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_2__[\"default\"])(\n  _TreeItem_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_1__[\"default\"],\n  _TreeItem_vue_vue_type_template_id_1c2a3381___WEBPACK_IMPORTED_MODULE_0__[\"render\"],\n  _TreeItem_vue_vue_type_template_id_1c2a3381___WEBPACK_IMPORTED_MODULE_0__[\"staticRenderFns\"],\n  false,\n  null,\n  null,\n  null\n  \n)\n\n/* hot reload */\nif (false) { var api; }\ncomponent.options.__file = \"src/components/TreeItem.vue\"\n/* harmony default export */ __webpack_exports__[\"default\"] = (component.exports);\n\n//# sourceURL=webpack:///./src/components/TreeItem.vue?");
	
	/***/ }),
	
	/***/ "./src/components/TreeItem.vue?vue&type=script&lang=ts&":
	/*!**************************************************************!*\
	  !*** ./src/components/TreeItem.vue?vue&type=script&lang=ts& ***!
	  \**************************************************************/
	/*! exports provided: default */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _node_modules_cache_loader_dist_cjs_js_ref_13_0_node_modules_ts_loader_index_js_ref_13_1_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_TreeItem_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../../node_modules/cache-loader/dist/cjs.js??ref--13-0!../../node_modules/ts-loader??ref--13-1!../../node_modules/cache-loader/dist/cjs.js??ref--0-0!../../node_modules/vue-loader/lib??vue-loader-options!./TreeItem.vue?vue&type=script&lang=ts& */ \"./node_modules/cache-loader/dist/cjs.js?!./node_modules/ts-loader/index.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/components/TreeItem.vue?vue&type=script&lang=ts&\");\n/* empty/unused harmony star reexport */ /* harmony default export */ __webpack_exports__[\"default\"] = (_node_modules_cache_loader_dist_cjs_js_ref_13_0_node_modules_ts_loader_index_js_ref_13_1_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_TreeItem_vue_vue_type_script_lang_ts___WEBPACK_IMPORTED_MODULE_0__[\"default\"]); \n\n//# sourceURL=webpack:///./src/components/TreeItem.vue?");
	
	/***/ }),
	
	/***/ "./src/components/TreeItem.vue?vue&type=template&id=1c2a3381&":
	/*!********************************************************************!*\
	  !*** ./src/components/TreeItem.vue?vue&type=template&id=1c2a3381& ***!
	  \********************************************************************/
	/*! exports provided: render, staticRenderFns */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _node_modules_cache_loader_dist_cjs_js_cacheDirectory_node_modules_cache_vue_loader_cacheIdentifier_3c385f58_vue_loader_template_node_modules_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_TreeItem_vue_vue_type_template_id_1c2a3381___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../../node_modules/cache-loader/dist/cjs.js?{\"cacheDirectory\":\"node_modules/.cache/vue-loader\",\"cacheIdentifier\":\"3c385f58-vue-loader-template\"}!../../node_modules/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!../../node_modules/cache-loader/dist/cjs.js??ref--0-0!../../node_modules/vue-loader/lib??vue-loader-options!./TreeItem.vue?vue&type=template&id=1c2a3381& */ \"./node_modules/cache-loader/dist/cjs.js?{\\\"cacheDirectory\\\":\\\"node_modules/.cache/vue-loader\\\",\\\"cacheIdentifier\\\":\\\"3c385f58-vue-loader-template\\\"}!./node_modules/vue-loader/lib/loaders/templateLoader.js?!./node_modules/cache-loader/dist/cjs.js?!./node_modules/vue-loader/lib/index.js?!./src/components/TreeItem.vue?vue&type=template&id=1c2a3381&\");\n/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, \"render\", function() { return _node_modules_cache_loader_dist_cjs_js_cacheDirectory_node_modules_cache_vue_loader_cacheIdentifier_3c385f58_vue_loader_template_node_modules_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_TreeItem_vue_vue_type_template_id_1c2a3381___WEBPACK_IMPORTED_MODULE_0__[\"render\"]; });\n\n/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, \"staticRenderFns\", function() { return _node_modules_cache_loader_dist_cjs_js_cacheDirectory_node_modules_cache_vue_loader_cacheIdentifier_3c385f58_vue_loader_template_node_modules_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_node_modules_cache_loader_dist_cjs_js_ref_0_0_node_modules_vue_loader_lib_index_js_vue_loader_options_TreeItem_vue_vue_type_template_id_1c2a3381___WEBPACK_IMPORTED_MODULE_0__[\"staticRenderFns\"]; });\n\n\n\n//# sourceURL=webpack:///./src/components/TreeItem.vue?");
	
	/***/ }),
	
	/***/ "./src/main.ts":
	/*!*********************!*\
	  !*** ./src/main.ts ***!
	  \*********************/
	/*! no exports provided */
	/***/ (function(module, __webpack_exports__, __webpack_require__) {
	
	"use strict";
	eval("__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var vue__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! vue */ \"./node_modules/vue/dist/vue.runtime.esm.js\");\n/* harmony import */ var _vue_composition_api__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @vue/composition-api */ \"./node_modules/@vue/composition-api/dist/vue-composition-api.module.js\");\n/* harmony import */ var _App_vue__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ./App.vue */ \"./src/App.vue\");\n\n\nvue__WEBPACK_IMPORTED_MODULE_0__[\"default\"].use(_vue_composition_api__WEBPACK_IMPORTED_MODULE_1__[\"default\"]);\n\nvue__WEBPACK_IMPORTED_MODULE_0__[\"default\"].config.productionTip = false;\nnew vue__WEBPACK_IMPORTED_MODULE_0__[\"default\"]({\n    render: h => h(_App_vue__WEBPACK_IMPORTED_MODULE_2__[\"default\"])\n}).$mount(\"#app\");\n\n\n//# sourceURL=webpack:///./src/main.ts?");
	
	/***/ }),
	
	/***/ 0:
	/*!***************************!*\
	  !*** multi ./src/main.ts ***!
	  \***************************/
	/*! no static exports found */
	/***/ (function(module, exports, __webpack_require__) {
	
	eval("module.exports = __webpack_require__(/*! ./src/main.ts */\"./src/main.ts\");\n\n\n//# sourceURL=webpack:///multi_./src/main.ts?");
	
	/***/ })
	
	/******/ });`
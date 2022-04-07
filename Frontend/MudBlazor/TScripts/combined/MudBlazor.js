/*!
* MudBlazor (https://mudblazor.com/)
* Copyright (c) 2021 MudBlazor
* Licensed under MIT (https://github.com/MudBlazor/MudBlazor/blob/master/LICENSE)
*/
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
window.mudDragAndDrop = {
initDropZone: (id) => {
const elem = document.getElementById('mud-drop-zone-' + id);
elem.addEventListener('dragover',() => event.preventDefault());
elem.addEventListener('dragstart', () => event.dataTransfer.setData('', event.target.id));
}
};
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudElementReference {
constructor() {
this.listenerId = 0;
this.eventListeners = {};
}
focus (element) {
if (element)
{
element.focus();
}
}
focusFirst (element, skip = 0, min = 0) {
if (element)
{
let tabbables = getTabbableElements(element);
if (tabbables.length <= min)
element.focus();
else
tabbables[skip].focus();
}
}
focusLast (element, skip = 0, min = 0) {
if (element)
{
let tabbables = getTabbableElements(element);
if (tabbables.length <= min)
element.focus();
else
tabbables[tabbables.length - skip - 1].focus();
}
}
saveFocus (element) {
if (element)
{
element['mudblazor_savedFocus'] = document.activeElement;
}
}
restoreFocus (element) {
if (element)
{
let previous = element['mudblazor_savedFocus'];
delete element['mudblazor_savedFocus']
if (previous)
previous.focus();
}
}
selectRange(element, pos1, pos2) {
if (element)
{
if (element.createTextRange) {
let selRange = element.createTextRange();
selRange.collapse(true);
selRange.moveStart('character', pos1);
selRange.moveEnd('character', pos2);
selRange.select();
} else if (element.setSelectionRange) {
element.setSelectionRange(pos1, pos2);
} else if (element.selectionStart) {
element.selectionStart = pos1;
element.selectionEnd = pos2;
}
element.focus();
}
}
select(element) {
if (element)
{
element.select();
}
}
/**
* gets the client rect of the parent of the element
* @param {HTMLElement} element
*/
getClientRectFromParent(element) {
if (!element) return;
let parent = element.parentElement;
if (!parent) return;
return this.getBoundingClientRect(parent);
}
/**
* Gets the client rect of the first child of the element
* @param {any} element
*/
getClientRectFromFirstChild(element) {
if (!element) return;
let child = element.children && element.children[0];
if (!child) return;
return this.getBoundingClientRect(child);
}
getBoundingClientRect(element) {
if (!element) return;
var rect = JSON.parse(JSON.stringify(element.getBoundingClientRect()));
rect.scrollY = window.scrollY || document.documentElement.scrollTop;
rect.scrollX = window.scrollX || document.documentElement.scrollLeft;
rect.windowHeight = window.innerHeight;
rect.windowWidth = window.innerWidth;
return rect;
}
/**
* Returns true if the element has any ancestor with style position==="fixed"
* @param {Element} element
*/
hasFixedAncestors(element) {
for (; element && element !== document; element = element.parentNode) {
if (window.getComputedStyle(element).getPropertyValue("position") === "fixed")
return true;
}
return false
};
changeCss (element, css) {
if (element)
{
element.className = css;
}
}
changeCssVariable (element, name, newValue) {
if (element)
{
element.style.setProperty(name, newValue);
}
}
addEventListener (element, dotnet, event, callback, spec, stopPropagation) {
let listener = function (e) {
const args = Array.from(spec, x => serializeParameter(e, x));
dotnet.invokeMethodAsync(callback, ...args);
if (stopPropagation) {
e.stopPropagation();
}
};
element.addEventListener(event, listener);
this.eventListeners[++this.listenerId] = listener;
return this.listenerId;
}
removeEventListener (element, event, eventId) {
element.removeEventListener(event, this.eventListeners[eventId]);
delete this.eventListeners[eventId];
}
};
window.mudElementRef = new MudElementReference();
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
//Functions related to MudThrottledEventManager
class MudThrottledEventManager {
constructor() {
this.mapper = {};
}
subscribe(eventName, elementId, projection, throotleInterval, key, properties, dotnetReference) {
const handlerRef = this.throttleEventHandler.bind(this, key);
let elem = document.getElementById(elementId);
if (elem) {
elem.addEventListener(eventName, handlerRef, false);
let projector = null;
if (projection) {
const parts = projection.split('.');
let functionPointer = window;
let functionReferenceFound = true;
if (parts.length == 0 || parts.length == 1) {
functionPointer = functionPointer[projection];
}
else {
for (let i = 0; i < parts.length; i++) {
functionPointer = functionPointer[parts[i]];
if (!functionPointer) {
functionReferenceFound = false;
break;
}
}
}
if (functionReferenceFound === true) {
projector = functionPointer;
}
}
this.mapper[key] = {
eventName: eventName,
handler: handlerRef,
delay: throotleInterval,
timerId: -1,
reference: dotnetReference,
elementId: elementId,
properties: properties,
projection: projector,
};
}
}
throttleEventHandler(key, event) {
const entry = this.mapper[key];
if (!entry) {
return;
}
clearTimeout(entry.timerId);
entry.timerId = window.setTimeout(
this.eventHandler.bind(this, key, event),
entry.delay
);
}
eventHandler(key, event) {
const entry = this.mapper[key];
if (!entry) {
return;
}
var elem = document.getElementById(entry.elementId);
if (elem != event.srcElement) {
return;
}
const eventEntry = {};
for (var i = 0; i < entry.properties.length; i++) {
eventEntry[entry.properties[i]] = event[entry.properties[i]];
}
if (entry.projection) {
if (typeof entry.projection === "function") {
entry.projection.apply(null, [eventEntry, event]);
}
}
entry.reference.invokeMethodAsync('OnEventOccur', key, JSON.stringify(eventEntry));
}
unsubscribe(key) {
const entry = this.mapper[key];
if (!entry) {
return;
}
entry.reference = null;
const elem = document.getElementById(entry.elementId);
if (elem) {
elem.removeEventListener(entry.eventName, entry.handler, false);
}
delete this.mapper[key];
}
};
window.mudThrottledEventManager = new MudThrottledEventManager();
window.mudEventProjections = {
correctOffset: function (eventEntry, event) {
var target = event.target.getBoundingClientRect();
eventEntry.offsetX = event.clientX - target.x;
eventEntry.offsetY = event.clientY - target.y;
}
};
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
window.getTabbableElements = (element) => {
return element.querySelectorAll(
"a[href]:not([tabindex='-1'])," +
"area[href]:not([tabindex='-1'])," +
"button:not([disabled]):not([tabindex='-1'])," +
"input:not([disabled]):not([tabindex='-1']):not([type='hidden'])," +
"select:not([disabled]):not([tabindex='-1'])," +
"textarea:not([disabled]):not([tabindex='-1'])," +
"iframe:not([tabindex='-1'])," +
"details:not([tabindex='-1'])," +
"[tabindex]:not([tabindex='-1'])," +
"[contentEditable=true]:not([tabindex='-1']"
);
};
//from: https://github.com/RemiBou/BrowserInterop
window.serializeParameter = (data, spec) => {
if (typeof data == "undefined" ||
data === null) {
return null;
}
if (typeof data === "number" ||
typeof data === "string" ||
typeof data == "boolean") {
return data;
}
let res = (Array.isArray(data)) ? [] : {};
if (!spec) {
spec = "*";
}
for (let i in data) {
let currentMember = data[i];
if (typeof currentMember === 'function' || currentMember === null) {
continue;
}
let currentMemberSpec;
if (spec != "*") {
currentMemberSpec = Array.isArray(data) ? spec : spec[i];
if (!currentMemberSpec) {
continue;
}
} else {
currentMemberSpec = "*"
}
if (typeof currentMember === 'object') {
if (Array.isArray(currentMember) || currentMember.length) {
res[i] = [];
for (let j = 0; j < currentMember.length; j++) {
const arrayItem = currentMember[j];
if (typeof arrayItem === 'object') {
res[i].push(this.serializeParameter(arrayItem, currentMemberSpec));
} else {
res[i].push(arrayItem);
}
}
} else {
//the browser provides some member (like plugins) as hash with index as key, if length == 0 we shall not convert it
if (currentMember.length === 0) {
res[i] = [];
} else {
res[i] = this.serializeParameter(currentMember, currentMemberSpec);
}
}
} else {
// string, number or boolean
if (currentMember === Infinity) { //inifity is not serialized by JSON.stringify
currentMember = "Infinity";
}
if (currentMember !== null) { //needed because the default json serializer in jsinterop serialize null values
res[i] = currentMember;
}
}
}
return res;
};
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudJsEventFactory {
connect(dotNetRef, elementId, options) {
//console.log('[MudBlazor | MudJsEventFactory] connect ', { dotNetRef, elementId, options });
if (!elementId)
throw "[MudBlazor | JsEvent] elementId: expected element id!";
var element = document.getElementById(elementId);
if (!element)
throw "[MudBlazor | JsEvent] no element found for id: " + elementId;
if (!element.mudJsEvent)
element.mudJsEvent = new MudJsEvent(dotNetRef, options);
element.mudJsEvent.connect(element);
}
disconnect(elementId) {
var element = document.getElementById(elementId);
if (!element || !element.mudJsEvent)
return;
element.mudJsEvent.disconnect();
}
subscribe(elementId, eventName) {
//console.log('[MudBlazor | MudJsEventFactory] subscribe ', { elementId, eventName});
if (!elementId)
throw "[MudBlazor | JsEvent] elementId: expected element id!";
var element = document.getElementById(elementId);
if (!element)
throw "[MudBlazor | JsEvent] no element found for id: " +elementId;
if (!element.mudJsEvent)
throw "[MudBlazor | JsEvent] please connect before subscribing"
element.mudJsEvent.subscribe(eventName);
}
unsubscribe(elementId, eventName) {
var element = document.getElementById(elementId);
if (!element || !element.mudJsEvent)
return;
element.mudJsEvent.unsubscribe(element, eventName);
}
}
window.mudJsEvent = new MudJsEventFactory();
class MudJsEvent {
constructor(dotNetRef, options) {
this._dotNetRef = dotNetRef;
this._options = options || {};
this.logger = options.enableLogging ? console.log : (message) => { };
this.logger('[MudBlazor | JsEvent] Initialized', { options });
this._subscribedEvents = {};
}
connect(element) {
if (!this._options)
return;
if (!this._options.targetClass)
throw "_options.targetClass: css class name expected";
if (this._observer) {
// don't do double registration
return;
}
var targetClass = this._options.targetClass;
this.logger('[MudBlazor | JsEvent] Start observing DOM of element for changes to child with class ', { element, targetClass });
this._element = element;
this._observer = new MutationObserver(this.onDomChanged);
this._observer.mudJsEvent = this;
this._observer.observe(this._element, { attributes: false, childList: true, subtree: true });
this._observedChildren = [];
}
disconnect() {
if (!this._observer)
return;
this.logger('[MudBlazor | JsEvent] disconnect mutation observer and event handler ');
this._observer.disconnect();
this._observer = null;
for (const child of this._observedChildren)
this.detachHandlers(child);
}
subscribe(eventName) {
// register handlers
if (this._subscribedEvents[eventName]) {
//console.log("... already attached");
return;
}
var element = this._element;
var targetClass = this._options.targetClass;
//this.logger('[MudBlazor | JsEvent] Subscribe event ' + eventName, { element, targetClass });
this._subscribedEvents[eventName]=true;
for (const child of element.getElementsByClassName(targetClass)) {
this.attachHandlers(child);
}
}
unsubscribe(eventName) {
if (!this._observer)
return;
this.logger('[MudBlazor | JsEvent] unsubscribe event handler ' + eventName );
this._observer.disconnect();
this._observer = null;
this._subscribedEvents[eventName] = false;
for (const child of this._observedChildren) {
this.detachHandler(child, eventName);
}
}
attachHandlers(child) {
child.mudJsEvent = this;
//this.logger('[MudBlazor | JsEvent] attachHandlers ', this._subscribedEvents, child);
for (var eventName of Object.getOwnPropertyNames(this._subscribedEvents)) {
if (!this._subscribedEvents[eventName])
continue;
// note: multiple registration of the same event not possible due to the use of the same handler func
this.logger('[MudBlazor | JsEvent] attaching event ' + eventName, child);
child.addEventListener(eventName, this.eventHandler);
}
if(this._observedChildren.indexOf(child) < 0)
this._observedChildren.push(child);
}
detachHandler(child, eventName) {
this.logger('[MudBlazor | JsEvent] detaching handler ' + eventName, child);
child.removeEventListener(eventName, this.eventHandler);
}
detachHandlers(child) {
this.logger('[MudBlazor | JsEvent] detaching handlers ', child);
for (var eventName of Object.getOwnPropertyNames(this._subscribedEvents)) {
if (!this._subscribedEvents[eventName])
continue;
child.removeEventListener(eventName, this.eventHandler);
}
this._observedChildren = this._observedChildren.filter(x=>x!==child);
}
onDomChanged(mutationsList, observer) {
var self = this.mudJsEvent; // func is invoked with this == _observer
//self.logger('[MudBlazor | JsEvent] onDomChanged: ', { self });
var targetClass = self._options.targetClass;
for (const mutation of mutationsList) {
//self.logger('[MudBlazor | JsEvent] Subtree mutation: ', { mutation });
for (const element of mutation.addedNodes) {
if (element.classList && element.classList.contains(targetClass)) {
if (!self._options.TagName || element.tagName == self._options.TagName)
self.attachHandlers(element);
}
}
for (const element of mutation.removedNodes) {
if (element.classList && element.classList.contains(targetClass)) {
if (!self._options.tagName || element.tagName == self._options.tagName)
self.detachHandlers(element);
}
}
}
}
eventHandler(e) {
var self = this.mudJsEvent; // func is invoked with this == child
var eventName = e.type;
self.logger('[MudBlazor | JsEvent] "' + eventName + '"', e);
// call specific handler
self["on" + eventName](self, e);
}
onkeyup(self, e) {
const caretPosition = e.target.selectionStart;
const invoke = self._subscribedEvents["keyup"];
if (invoke) {
//self.logger('[MudBlazor | JsEvent] caret pos: ' + caretPosition);
self._dotNetRef.invokeMethodAsync('OnCaretPositionChanged', caretPosition);
}
}
onclick(self, e) {
const caretPosition = e.target.selectionStart;
const invoke = self._subscribedEvents["click"];
if (invoke) {
//self.logger('[MudBlazor | JsEvent] caret pos: ' + caretPosition);
self._dotNetRef.invokeMethodAsync('OnCaretPositionChanged', caretPosition);
}
}
//oncopy(self, e) {
//    const invoke = self._subscribedEvents["copy"];
//    if (invoke) {
//        //self.logger('[MudBlazor | JsEvent] copy (preventing default and stopping propagation)');
//        e.preventDefault();
//        e.stopPropagation();
//        self._dotNetRef.invokeMethodAsync('OnCopy');
//    }
//}
onpaste(self, e) {
const invoke = self._subscribedEvents["paste"];
if (invoke) {
//self.logger('[MudBlazor | JsEvent] paste (preventing default and stopping propagation)');
e.preventDefault();
e.stopPropagation();
const text = (e.originalEvent || e).clipboardData.getData('text/plain');
self._dotNetRef.invokeMethodAsync('OnPaste', text);
}
}
onselect(self, e) {
const invoke = self._subscribedEvents["select"];
if (invoke) {
const start = e.target.selectionStart;
const end = e.target.selectionEnd;
if (start === end)
return; // <-- we have caret position changed for that.
//self.logger('[MudBlazor | JsEvent] select ' + start + "-" + end);
self._dotNetRef.invokeMethodAsync('OnSelect', start, end);
}
}
}
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudKeyInterceptorFactory {
connect(dotNetRef, elementId, options) {
//console.log('[MudBlazor | MudKeyInterceptorFactory] connect ', { dotNetRef, element, options });
if (!elementId)
throw "elementId: expected element id!";
var element = document.getElementById(elementId);
if (!element)
throw "no element found for id: " +elementId;
if (!element.mudKeyInterceptor)
element.mudKeyInterceptor = new MudKeyInterceptor(dotNetRef, options);
element.mudKeyInterceptor.connect(element);
}
updatekey(elementId, option) {
var element = document.getElementById(elementId);
if (!element || !element.mudKeyInterceptor)
return;
element.mudKeyInterceptor.updatekey(option);
}
disconnect(elementId) {
var element = document.getElementById(elementId);
if (!element || !element.mudKeyInterceptor)
return;
element.mudKeyInterceptor.disconnect();
}
}
window.mudKeyInterceptor = new MudKeyInterceptorFactory();
class MudKeyInterceptor {
constructor(dotNetRef, options) {
this._dotNetRef = dotNetRef;
this._options = options;
this.logger = options.enableLogging ? console.log : (message) => { };
this.logger('[MudBlazor | KeyInterceptor] Interceptor initialized', { options });
}
connect(element) {
if (!this._options)
return;
if (!this._options.keys)
throw "_options.keys: array of KeyOptions expected";
if (!this._options.targetClass)
throw "_options.targetClass: css class name expected";
if (this._observer) {
// don't do double registration
return;
}
var targetClass = this._options.targetClass;
this.logger('[MudBlazor | KeyInterceptor] Start observing DOM of element for changes to child with class ', { element, targetClass});
this._element = element;
this._observer = new MutationObserver(this.onDomChanged);
this._observer.mudKeyInterceptor = this;
this._observer.observe(this._element, { attributes: false, childList: true, subtree: true });
this._observedChildren = [];
// transform key options into a key lookup
this._keyOptions = {};
this._regexOptions = [];
for (const keyOption of this._options.keys) {
if (!keyOption || !keyOption.key) {
this.logger('[MudBlazor | KeyInterceptor] got invalid key options: ', keyOption);
continue;
}
this.setKeyOption(keyOption)
}
this.logger('[MudBlazor | KeyInterceptor] key options: ', this._keyOptions);
if (this._regexOptions.size > 0)
this.logger('[MudBlazor | KeyInterceptor] regex options: ', this._regexOptions);
// register handlers
for (const child of this._element.getElementsByClassName(targetClass)) {
this.attachHandlers(child);
}
}
setKeyOption(keyOption) {
if (keyOption.key.length > 2 && keyOption.key.startsWith('/') && keyOption.key.endsWith('/')) {
// JS regex key options such as "/[a-z]/" or "/a|b/" but NOT "/[a-z]/g" or "/[a-z]/i"
keyOption.regex = new RegExp(keyOption.key.substring(1, keyOption.key.length - 1)); // strip the / from start and end
this._regexOptions.push(keyOption);
}
else
this._keyOptions[keyOption.key.toLowerCase()] = keyOption;
// remove whitespace and enforce lowercase
var whitespace = new RegExp("\\s", "g");
keyOption.preventDown = (keyOption.preventDown || "none").replace(whitespace, "").toLowerCase();
keyOption.preventUp = (keyOption.preventUp || "none").replace(whitespace, "").toLowerCase();
keyOption.stopDown = (keyOption.stopDown || "none").replace(whitespace, "").toLowerCase();
keyOption.stopUp = (keyOption.stopUp || "none").replace(whitespace, "").toLowerCase();
}
updatekey(updatedOption) {
var option = this._keyOptions[updatedOption.key.toLowerCase()];
option || this.logger('[MudBlazor | KeyInterceptor] updating option failed: key not registered');
this.setKeyOption(updatedOption);
this.logger('[MudBlazor | KeyInterceptor] updated option ', { option, updatedOption });
}
disconnect() {
if (!this._observer)
return;
this.logger('[MudBlazor | KeyInterceptor] disconnect mutation observer and event handlers');
this._observer.disconnect();
this._observer = null;
for (const child of this._observedChildren)
this.detachHandlers(child);
}
attachHandlers(child) {
this.logger('[MudBlazor | KeyInterceptor] attaching handlers ', { child });
if (this._observedChildren.indexOf(child) > -1) {
//console.log("... already attached");
return;
}
child.mudKeyInterceptor = this;
child.addEventListener('keydown', this.onKeyDown);
child.addEventListener('keyup', this.onKeyUp);
this._observedChildren.push(child);
}
detachHandlers(child) {
this.logger('[MudBlazor | KeyInterceptor] detaching handlers ', { child });
child.removeEventListener('keydown', this.onKeyDown);
child.removeEventListener('keyup', this.onKeyUp);
this._observedChildren = this._observedChildren.filter(x=>x!==child);
}
onDomChanged(mutationsList, observer) {
var self = this.mudKeyInterceptor; // func is invoked with this == _observer
//self.logger('[MudBlazor | KeyInterceptor] onDomChanged: ', { self });
var targetClass = self._options.targetClass;
for (const mutation of mutationsList) {
//self.logger('[MudBlazor | KeyInterceptor] Subtree mutation: ', { mutation });
for (const element of mutation.addedNodes) {
if (element.classList && element.classList.contains(targetClass))
self.attachHandlers(element);
}
for (const element of mutation.removedNodes) {
if (element.classList && element.classList.contains(targetClass))
self.detachHandlers(element);
}
}
}
matchesKeyCombination(option, args) {
if (!option || option=== "none")
return false;
if (option === "any")
return true;
var shift = args.shiftKey;
var ctrl = args.ctrlKey;
var alt = args.altKey;
var meta = args.metaKey;
var any = shift || ctrl || alt || meta;
if (any && option === "key+any")
return true;
if (!any && option.includes("key+none"))
return true;
if (!any)
return false;
var combi = `key${shift ? "+shift" : ""}${ctrl ? "+ctrl" : ""}${alt ? "+alt" : ""}${meta ? "+meta" : ""}`;
return option.includes(combi);
}
onKeyDown(args) {
var self = this.mudKeyInterceptor; // func is invoked with this == child
var key = args.key.toLowerCase();
self.logger('[MudBlazor | KeyInterceptor] down "' + key + '"', args);
var invoke = false;
if (self._keyOptions.hasOwnProperty(key)) {
var keyOptions = self._keyOptions[key];
self.logger('[MudBlazor | KeyInterceptor] options for "' + key + '"', keyOptions);
self.processKeyDown(args, keyOptions);
if (keyOptions.subscribeDown)
invoke = true;
}
for (const keyOptions of self._regexOptions) {
if (keyOptions.regex.test(key)) {
self.logger('[MudBlazor | KeyInterceptor] regex options for "' + key + '"', keyOptions);
self.processKeyDown(args, keyOptions);
if (keyOptions.subscribeDown)
invoke = true;
}
}
if (invoke) {
var eventArgs = self.toKeyboardEventArgs(args);
eventArgs.Type = "keydown";
// we'd like to pass a reference to the child element back to dotnet but we can't
// https://github.com/dotnet/aspnetcore/issues/16110
// if we ever need it we'll pass the id up and users need to id the observed elements
self._dotNetRef.invokeMethodAsync('OnKeyDown', eventArgs);
}
}
processKeyDown(args, keyOptions) {
if (this.matchesKeyCombination(keyOptions.preventDown, args))
args.preventDefault();
if (this.matchesKeyCombination(keyOptions.stopDown, args))
args.stopPropagation();
}
onKeyUp(args) {
var self = this.mudKeyInterceptor; // func is invoked with this == child
var key = args.key.toLowerCase();
self.logger('[MudBlazor | KeyInterceptor] up "' + key + '"', args);
var invoke = false;
if (self._keyOptions.hasOwnProperty(key)) {
var keyOptions = self._keyOptions[key];
self.processKeyUp(args, keyOptions);
if (keyOptions.subscribeUp)
invoke = true;
}
for (const keyOptions of self._regexOptions) {
if (keyOptions.regex.test(key)) {
self.processKeyUp(args, keyOptions);
if (keyOptions.subscribeUp)
invoke = true;
}
}
if (invoke) {
var eventArgs = self.toKeyboardEventArgs(args);
eventArgs.Type = "keyup";
// we'd like to pass a reference to the child element back to dotnet but we can't
// https://github.com/dotnet/aspnetcore/issues/16110
// if we ever need it we'll pass the id up and users need to id the observed elements
self._dotNetRef.invokeMethodAsync('OnKeyUp', eventArgs);
}
}
processKeyUp(args, keyOptions) {
if (this.matchesKeyCombination(keyOptions.preventUp, args))
args.preventDefault();
if (this.matchesKeyCombination(keyOptions.stopUp, args))
args.stopPropagation();
}
toKeyboardEventArgs(args) {
return {
Key: args.key,
Code: args.code,
Location: args.location,
Repeat: args.repeat,
CtrlKey: args.ctrlKey,
ShiftKey: args.shiftKey,
AltKey: args.altKey,
MetaKey: args.metaKey
};
}
}
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
window.mudpopoverHelper = {
calculatePopoverPosition: function (list, boundingRect, selfRect) {
let top = 0;
let left = 0;
if (list.indexOf('mud-popover-anchor-top-left') >= 0) {
left = boundingRect.left;
top = boundingRect.top;
} else if (list.indexOf('mud-popover-anchor-top-center') >= 0) {
left = boundingRect.left + boundingRect.width / 2;
top = boundingRect.top;
} else if (list.indexOf('mud-popover-anchor-top-right') >= 0) {
left = boundingRect.left + boundingRect.width;
top = boundingRect.top;
} else if (list.indexOf('mud-popover-anchor-center-left') >= 0) {
left = boundingRect.left;
top = boundingRect.top + boundingRect.height / 2;
} else if (list.indexOf('mud-popover-anchor-center-center') >= 0) {
left = boundingRect.left + boundingRect.width / 2;
top = boundingRect.top + boundingRect.height / 2;
} else if (list.indexOf('mud-popover-anchor-center-right') >= 0) {
left = boundingRect.left + boundingRect.width;
top = boundingRect.top + boundingRect.height / 2;
} else if (list.indexOf('mud-popover-anchor-bottom-left') >= 0) {
left = boundingRect.left;
top = boundingRect.top + boundingRect.height;
} else if (list.indexOf('mud-popover-anchor-bottom-center') >= 0) {
left = boundingRect.left + boundingRect.width / 2;
top = boundingRect.top + boundingRect.height;
} else if (list.indexOf('mud-popover-anchor-bottom-right') >= 0) {
left = boundingRect.left + boundingRect.width;
top = boundingRect.top + boundingRect.height;
}
let offsetX = 0;
let offsetY = 0;
if (list.indexOf('mud-popover-top-left') >= 0) {
offsetX = 0;
offsetY = 0;
} else if (list.indexOf('mud-popover-top-center') >= 0) {
offsetX = -selfRect.width / 2;
offsetY = 0;
} else if (list.indexOf('mud-popover-top-right') >= 0) {
offsetX = -selfRect.width;
offsetY = 0;
}
else if (list.indexOf('mud-popover-center-left') >= 0) {
offsetX = 0;
offsetY = -selfRect.height / 2;
} else if (list.indexOf('mud-popover-center-center') >= 0) {
offsetX = -selfRect.width / 2;
offsetY = -selfRect.height / 2;
} else if (list.indexOf('mud-popover-center-right') >= 0) {
offsetX = -selfRect.width;
offsetY = -selfRect.height / 2;
}
else if (list.indexOf('mud-popover-bottom-left') >= 0) {
offsetX = 0;
offsetY = -selfRect.height;
} else if (list.indexOf('mud-popover-bottom-center') >= 0) {
offsetX = -selfRect.width / 2;
offsetY = -selfRect.height;
} else if (list.indexOf('mud-popover-bottom-right') >= 0) {
offsetX = -selfRect.width;
offsetY = -selfRect.height;
}
return {
top: top, left: left, offsetX: offsetX, offsetY: offsetY
};
},
flipClassReplacements: {
'top': {
'mud-popover-top-left': 'mud-popover-bottom-left',
'mud-popover-top-center': 'mud-popover-bottom-center',
'mud-popover-anchor-bottom-center': 'mud-popover-anchor-top-center',
'mud-popover-top-right': 'mud-popover-bottom-right',
},
'left': {
'mud-popover-top-left': 'mud-popover-top-right',
'mud-popover-center-left': 'mud-popover-center-right',
'mud-popover-anchor-center-right': 'mud-popover-anchor-center-left',
'mud-popover-bottom-left': 'mud-popover-bottom-right',
},
'right': {
'mud-popover-top-right': 'mud-popover-top-left',
'mud-popover-center-right': 'mud-popover-center-left',
'mud-popover-anchor-center-left': 'mud-popover-anchor-center-right',
'mud-popover-bottom-right': 'mud-popover-bottom-left',
},
'bottom': {
'mud-popover-bottom-left': 'mud-popover-top-left',
'mud-popover-bottom-center': 'mud-popover-top-center',
'mud-popover-anchor-top-center': 'mud-popover-anchor-bottom-center',
'mud-popover-bottom-right': 'mud-popover-top-right',
},
'top-and-left': {
'mud-popover-top-left': 'mud-popover-bottom-right',
},
'top-and-right': {
'mud-popover-top-right': 'mud-popover-bottom-left',
},
'bottom-and-left': {
'mud-popover-bottom-left': 'mud-popover-top-right',
},
'bottom-and-right': {
'mud-popover-bottom-right': 'mud-popover-top-left',
},
},
flipMargin: 0,
getPositionForFlippedPopver: function (inputArray, selector, boundingRect, selfRect) {
const classList = [];
for (var i = 0; i < inputArray.length; i++) {
const item = inputArray[i];
const replacments = window.mudpopoverHelper.flipClassReplacements[selector][item];
if (replacments) {
classList.push(replacments);
}
else {
classList.push(item);
}
}
return window.mudpopoverHelper.calculatePopoverPosition(classList, boundingRect, selfRect);
},
placePopover: function (popoverNode, classSelector) {
if (popoverNode && popoverNode.parentNode) {
const id = popoverNode.id.substr(8);
const popoverContentNode = document.getElementById('popovercontent-' + id);
if (popoverContentNode.classList.contains('mud-popover-open') == false) {
return;
}
if (!popoverContentNode) {
return;
}
if (classSelector) {
if (popoverContentNode.classList.contains(classSelector) == false) {
return;
}
}
const boundingRect = popoverNode.parentNode.getBoundingClientRect();
if (popoverContentNode.classList.contains('mud-popover-relative-width')) {
popoverContentNode.style['max-width'] = (boundingRect.width) + 'px';
}
const selfRect = popoverContentNode.getBoundingClientRect();
const classList = popoverContentNode.classList;
const classListArray = Array.from(popoverContentNode.classList);
const postion = window.mudpopoverHelper.calculatePopoverPosition(classListArray, boundingRect, selfRect);
let left = postion.left;
let top = postion.top;
let offsetX = postion.offsetX;
let offsetY = postion.offsetY;
if (classList.contains('mud-popover-overflow-flip-onopen') || classList.contains('mud-popover-overflow-flip-always')) {
const appBarElements = document.getElementsByClassName("mud-appbar mud-appbar-fixed-top");
let appBarOffset = 0;
if (appBarElements.length > 0) {
appBarOffset = appBarElements[0].getBoundingClientRect().height;
}
const graceMargin = window.mudpopoverHelper.flipMargin;
const deltaToLeft = left + offsetX;
const deltaToRight = window.innerWidth - left - selfRect.width;
const deltaTop = top - selfRect.height - appBarOffset;
const spaceToTop = top - appBarOffset;
const deltaBottom = window.innerHeight - top - selfRect.height;
//console.log('self-width: ' + selfRect.width + ' | self-height: ' + selfRect.height);
//console.log('left: ' + deltaToLeft + ' | rigth:' + deltaToRight + ' | top: ' + deltaTop + ' | bottom: ' + deltaBottom + ' | spaceToTop: ' + spaceToTop);
let selector = popoverContentNode.mudPopoverFliped;
if (!selector) {
if (classList.contains('mud-popover-top-left')) {
if (deltaBottom < graceMargin && deltaToRight < graceMargin && spaceToTop >= selfRect.height && deltaToLeft >= selfRect.width) {
selector = 'top-and-left';
} else if (deltaBottom < graceMargin && spaceToTop >= selfRect.height) {
selector = 'top';
} else if (deltaToRight < graceMargin && deltaToLeft >= selfRect.width) {
selector = 'left';
}
} else if (classList.contains('mud-popover-top-center')) {
if (deltaBottom < graceMargin && spaceToTop >= selfRect.height) {
selector = 'top';
}
} else if (classList.contains('mud-popover-top-right')) {
if (deltaBottom < graceMargin && deltaToLeft < graceMargin && spaceToTop >= selfRect.height && deltaToRight >= selfRect.width) {
selector = 'top-and-right';
} else if (deltaBottom < graceMargin && spaceToTop >= selfRect.height) {
selector = 'top';
} else if (deltaToLeft < graceMargin && deltaToRight >= selfRect.width) {
selector = 'right';
}
}
else if (classList.contains('mud-popover-center-left')) {
if (deltaToRight < graceMargin && deltaToLeft >= selfRect.width) {
selector = 'left';
}
}
else if (classList.contains('mud-popover-center-right')) {
if (deltaToLeft < graceMargin && deltaToRight >= selfRect.width) {
selector = 'right';
}
}
else if (classList.contains('mud-popover-bottom-left')) {
if (deltaTop < graceMargin && deltaToRight < graceMargin && deltaBottom >= 0 && deltaToLeft >= selfRect.width) {
selector = 'bottom-and-left';
} else if (deltaTop < graceMargin && deltaBottom >= 0) {
selector = 'bottom';
} else if (deltaToRight < graceMargin && deltaToLeft >= selfRect.width) {
selector = 'left';
}
} else if (classList.contains('mud-popover-bottom-center')) {
if (deltaTop < graceMargin && deltaBottom >= 0) {
selector = 'bottom';
}
} else if (classList.contains('mud-popover-bottom-right')) {
if (deltaTop < graceMargin && deltaToLeft < graceMargin && deltaBottom >= 0 && deltaToRight >= selfRect.width) {
selector = 'bottom-and-right';
} else if (deltaTop < graceMargin && deltaBottom >= 0) {
selector = 'bottom';
} else if (deltaToLeft < graceMargin && deltaToRight >= selfRect.width) {
selector = 'right';
}
}
}
if (selector && selector != 'none') {
const newPosition = window.mudpopoverHelper.getPositionForFlippedPopver(classListArray, selector, boundingRect, selfRect);
left = newPosition.left;
top = newPosition.top;
offsetX = newPosition.offsetX;
offsetY = newPosition.offsetY;
popoverContentNode.setAttribute('data-mudpopover-flip', 'flipped');
}
else {
popoverContentNode.removeAttribute('data-mudpopover-flip');
}
if (classList.contains('mud-popover-overflow-flip-onopen')) {
if (!popoverContentNode.mudPopoverFliped) {
popoverContentNode.mudPopoverFliped = selector || 'none';
}
}
}
if (popoverContentNode.classList.contains('mud-popover-fixed')) {
}
else if (window.getComputedStyle(popoverNode).position == 'fixed') {
popoverContentNode.style['position'] = 'fixed';
}
else {
offsetX += window.scrollX;
offsetY += window.scrollY
}
popoverContentNode.style['left'] = (left + offsetX) + 'px';
popoverContentNode.style['top'] = (top + offsetY) + 'px';
if (window.getComputedStyle(popoverNode).getPropertyValue('z-index') != 'auto') {
popoverContentNode.style['z-index'] = window.getComputedStyle(popoverNode).getPropertyValue('z-index');
}
}
},
placePopoverByClassSelector: function (classSelector = null) {
var items = window.mudPopover.getAllObservedContainers();
for (let i = 0; i < items.length; i++) {
const popoverNode = document.getElementById('popover-' + items[i]);
window.mudpopoverHelper.placePopover(popoverNode, classSelector);
}
},
placePopoverByNode: function (target) {
const id = target.id.substr(15);
const popoverNode = document.getElementById('popover-' + id);
window.mudpopoverHelper.placePopover(popoverNode);
}
}
class MudPopover {
constructor() {
this.map = {};
this.contentObserver = null;
this.mainContainerClass = null;
}
callback(id, mutationsList, observer) {
for (const mutation of mutationsList) {
if (mutation.type === 'attributes') {
const target = mutation.target
if (target.classList.contains('mud-popover-overflow-flip-onopen') &&
target.classList.contains('mud-popover-open') == false) {
target.mudPopoverFliped = null;
target.removeAttribute('data-mudpopover-flip');
}
window.mudpopoverHelper.placePopoverByNode(target);
}
}
}
initilize(containerClass, flipMargin) {
const mainContent = document.getElementsByClassName(containerClass);
if (mainContent.length == 0) {
return;
}
if (flipMargin) {
window.mudpopoverHelper.flipMargin = flipMargin;
}
this.mainContainerClass = containerClass;
if (!mainContent[0].mudPopoverMark) {
mainContent[0].mudPopoverMark = "mudded";
if (this.contentObserver != null) {
this.contentObserver.disconnect();
this.contentObserver = null;
}
this.contentObserver = new ResizeObserver(entries => {
window.mudpopoverHelper.placePopoverByClassSelector();
});
this.contentObserver.observe(mainContent[0]);
}
}
connect(id) {
this.initilize(this.mainContainerClass);
const popoverNode = document.getElementById('popover-' + id);
const popoverContentNode = document.getElementById('popovercontent-' + id);
if (popoverNode && popoverNode.parentNode && popoverContentNode) {
window.mudpopoverHelper.placePopover(popoverNode);
const config = { attributeFilter: ['class'] };
const observer = new MutationObserver(this.callback.bind(this, id));
observer.observe(popoverContentNode, config);
const resizeObserver = new ResizeObserver(entries => {
for (let entry of entries) {
const target = entry.target;
for (var i = 0; i < target.childNodes.length; i++) {
const childNode = target.childNodes[i];
if (childNode.id && childNode.id.startsWith('popover-')) {
window.mudpopoverHelper.placePopover(childNode);
}
}
}
});
resizeObserver.observe(popoverNode.parentNode);
const contentNodeObserver = new ResizeObserver(entries => {
for (let entry of entries) {
var target = entry.target;
window.mudpopoverHelper.placePopoverByNode(target);
}
});
contentNodeObserver.observe(popoverContentNode);
this.map[id] = {
mutationObserver: observer,
resizeObserver: resizeObserver,
contentNodeObserver: contentNodeObserver
};
}
}
disconnect(id) {
if (this.map[id]) {
const item = this.map[id]
item.mutationObserver.disconnect();
item.resizeObserver.disconnect();
item.contentNodeObserver.disconnect();
delete this.map[id];
}
}
dispose() {
for (var i in this.map) {
disconnect(i);
}
this.contentObserver.disconnect();
this.contentObserver = null;
}
getAllObservedContainers() {
const result = [];
for (var i in this.map) {
result.push(i);
}
return result;
}
}
window.mudPopover = new MudPopover();
window.addEventListener('scroll', () => {
window.mudpopoverHelper.placePopoverByClassSelector('mud-popover-fixed');
window.mudpopoverHelper.placePopoverByClassSelector('mud-popover-overflow-flip-always');
});
window.addEventListener('resize', () => {
window.mudpopoverHelper.placePopoverByClassSelector();
});
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudResizeListener {
constructor(id) {
this.logger = function (message) { };
this.options = {};
this.throttleResizeHandlerId = -1;
this.dotnet = undefined;
this.breakpoint = -1;
this.id = id;
}
listenForResize(dotnetRef, options) {
if (this.dotnet) {
this.options = options;
return;
}
//this.logger("[MudBlazor] listenForResize:", { options, dotnetRef });
this.options = options;
this.dotnet = dotnetRef;
this.logger = options.enableLogging ? console.log : (message) => { };
this.logger(`[MudBlazor] Reporting resize events at rate of: ${(this.options || {}).reportRate || 100}ms`);
window.addEventListener("resize", this.throttleResizeHandler.bind(this), false);
if (!this.options.suppressInitEvent) {
this.resizeHandler();
}
this.breakpoint = this.getBreakpoint(window.innerWidth);
}
throttleResizeHandler() {
clearTimeout(this.throttleResizeHandlerId);
//console.log("[MudBlazor] throttleResizeHandler ", {options:this.options});
this.throttleResizeHandlerId = window.setTimeout(this.resizeHandler.bind(this), ((this.options || {}).reportRate || 100));
}
resizeHandler() {
if (this.options.notifyOnBreakpointOnly) {
let bp = this.getBreakpoint(window.innerWidth);
if (bp == this.breakpoint) {
return;
}
this.breakpoint = bp;
}
try {
//console.log("[MudBlazor] RaiseOnResized invoked");
if (this.id) {
this.dotnet.invokeMethodAsync('RaiseOnResized',
{
height: window.innerHeight,
width: window.innerWidth
},
this.getBreakpoint(window.innerWidth),
this.id);
}
else {
this.dotnet.invokeMethodAsync('RaiseOnResized',
{
height: window.innerHeight,
width: window.innerWidth
},
this.getBreakpoint(window.innerWidth));
}
//this.logger("[MudBlazor] RaiseOnResized invoked");
} catch (error) {
this.logger("[MudBlazor] Error in resizeHandler:", { error });
}
}
cancelListener() {
this.dotnet = undefined;
//console.log("[MudBlazor] cancelListener");
window.removeEventListener("resize", this.throttleResizeHandler);
}
matchMedia(query) {
let m = window.matchMedia(query).matches;
//this.logger(`[MudBlazor] matchMedia "${query}": ${m}`);
return m;
}
getBrowserWindowSize() {
//this.logger("[MudBlazor] getBrowserWindowSize");
return {
height: window.innerHeight,
width: window.innerWidth
};
}
getBreakpoint(width) {
if (width >= this.options.breakpointDefinitions["Xl"])
return 4;
else if (width >= this.options.breakpointDefinitions["Lg"])
return 3;
else if (width >= this.options.breakpointDefinitions["Md"])
return 2;
else if (width >= this.options.breakpointDefinitions["Sm"])
return 1;
else //Xs
return 0;
}
};
window.mudResizeListener = new MudResizeListener();
window.mudResizeListenerFactory = {
mapping: {},
listenForResize: (dotnetRef, options, id) => {
var map = window.mudResizeListenerFactory.mapping;
if (map[id]) {
return;
}
var listener = new MudResizeListener(id);
listener.listenForResize(dotnetRef, options);
map[id] = listener;
},
cancelListener: (id) => {
var map = window.mudResizeListenerFactory.mapping;
if (!map[id]) {
return;
}
var listener = map[id];
listener.cancelListener();
delete map[id];
},
cancelListeners: (ids) => {
for (let i = 0; i < ids.length; i++) {
window.mudResizeListenerFactory.cancelListener(ids[i]);
}
}
}
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudResizeObserverFactory {
constructor() {
this._maps = {};
}
connect(id, dotNetRef, elements, elementIds, options) {
var existingEntry = this._maps[id];
if (!existingEntry) {
var observer = new MudResizeObserver(dotNetRef, options);
this._maps[id] = observer;
}
var result = this._maps[id].connect(elements, elementIds);
return result;
}
disconnect(id, element) {
//I can't think about a case, where this can be called, without observe has been called before
//however, a check is not harmful either
var existingEntry = this._maps[id];
if (existingEntry) {
existingEntry.disconnect(element);
}
}
cancelListener(id) {
//cancelListener is called during dispose of .net instance
//in rare cases it could be possible, that no object has been connect so far
//and no entry exists. Therefore, a little check to prevent an error in this case
var existingEntry = this._maps[id];
if (existingEntry) {
existingEntry.cancelListener();
delete this._maps[id];
}
}
}
class MudResizeObserver {
constructor(dotNetRef, options) {
this.logger = options.enableLogging ? console.log : (message) => { };
this.options = options;
this._dotNetRef = dotNetRef
var delay = (this.options || {}).reportRate || 200;
this.throttleResizeHandlerId = -1;
var observervedElements = [];
this._observervedElements = observervedElements;
this.logger('[MudBlazor | ResizeObserver] Observer initilized');
this._resizeObserver = new ResizeObserver(entries => {
var changes = [];
this.logger('[MudBlazor | ResizeObserver] changes detected');
for (let entry of entries) {
var target = entry.target;
var affectedObservedElement = observervedElements.find((x) => x.element == target);
if (affectedObservedElement) {
var size = entry.target.getBoundingClientRect();
if (affectedObservedElement.isInitilized == true) {
changes.push({ id: affectedObservedElement.id, size: size });
}
else {
affectedObservedElement.isInitilized = true;
}
}
}
if (changes.length > 0) {
if (this.throttleResizeHandlerId >= 0) {
clearTimeout(this.throttleResizeHandlerId);
}
this.throttleResizeHandlerId = window.setTimeout(this.resizeHandler.bind(this, changes), delay);
}
});
}
resizeHandler(changes) {
try {
this.logger("[MudBlazor | ResizeObserver] OnSizeChanged handler invoked");
this._dotNetRef.invokeMethodAsync("OnSizeChanged", changes);
} catch (error) {
this.logger("[MudBlazor | ResizeObserver] Error in OnSizeChanged handler:", { error });
}
}
connect(elements, ids) {
var result = [];
this.logger('[MudBlazor | ResizeObserver] Start observing elements...');
for (var i = 0; i < elements.length; i++) {
var newEntry = {
element: elements[i],
id: ids[i],
isInitilized: false,
};
this.logger("[MudBlazor | ResizeObserver] Start observing element:", { newEntry });
result.push(elements[i].getBoundingClientRect());
this._observervedElements.push(newEntry);
this._resizeObserver.observe(elements[i]);
}
return result;
}
disconnect(elementId) {
this.logger('[MudBlazor | ResizeObserver] Try to unobserve element with id', { elementId });
var affectedObservedElement = this._observervedElements.find((x) => x.id == elementId);
if (affectedObservedElement) {
var element = affectedObservedElement.element;
this._resizeObserver.unobserve(element);
this.logger('[MudBlazor | ResizeObserver] Element found. Ubobserving size changes of element', { element });
var index = this._observervedElements.indexOf(affectedObservedElement);
this._observervedElements.splice(index, 1);
}
}
cancelListener() {
this.logger('[MudBlazor | ResizeObserver] Closing ResizeObserver. Detaching all observed elements');
this._resizeObserver.disconnect();
this._dotNetRef = undefined;
}
}
window.mudResizeObserver = new MudResizeObserverFactory();
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
//Functions related to scroll events
class MudScrollListener {
constructor() {
this.throttleScrollHandlerId = -1;
}
// subscribe to throttled scroll event
listenForScroll(dotnetReference, selector) {
//if selector is null, attach to document
let element = selector
? document.querySelector(selector)
: document;
// add the event listener
element.addEventListener(
'scroll',
this.throttleScrollHandler.bind(this, dotnetReference),
false
);
}
// fire the event just once each 100 ms, **it's hardcoded**
throttleScrollHandler(dotnetReference, event) {
clearTimeout(this.throttleScrollHandlerId);
this.throttleScrollHandlerId = window.setTimeout(
this.scrollHandler.bind(this, dotnetReference, event),
100
);
}
// when scroll event is fired, pass this information to
// the RaiseOnScroll C# method of the ScrollListener
// We pass the scroll coordinates of the element and
// the boundingClientRect of the first child, because
// scrollTop of body is always 0. With this information,
// we can trigger C# events on different scroll situations
scrollHandler(dotnetReference, event) {
try {
let element = event.target;
//data to pass
let scrollTop = element.scrollTop;
let scrollHeight = element.scrollHeight;
let scrollWidth = element.scrollWidth;
let scrollLeft = element.scrollLeft;
let nodeName = element.nodeName;
//data to pass
let firstChild = element.firstElementChild;
let firstChildBoundingClientRect = firstChild.getBoundingClientRect();
//invoke C# method
dotnetReference.invokeMethodAsync('RaiseOnScroll', {
firstChildBoundingClientRect,
scrollLeft,
scrollTop,
scrollHeight,
scrollWidth,
nodeName,
});
} catch (error) {
console.log('[MudBlazor] Error in scrollHandler:', { error });
}
}
//remove event listener
cancelListener(selector) {
let element = selector
? document.querySelector(selector)
: document.documentElement;
element.removeEventListener('scroll', this.throttleScrollHandler);
}
};
window.mudScrollListener = new MudScrollListener();
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudScrollManager {
//scrolls to an Id. Useful for navigation to fragments
scrollToFragment(elementId, behavior) {
let element = document.getElementById(elementId);
if (element) {
element.scrollIntoView({ behavior, block: 'center', inline: 'start' });
}
}
//scrolls to year in MudDatePicker
scrollToYear(elementId, offset) {
let element = document.getElementById(elementId);
if (element) {
element.parentNode.scrollTop = element.offsetTop - element.parentNode.offsetTop - element.scrollHeight * 3;
}
}
// sets the scroll position of the elements container,
// to the position of the element with the given element id
scrollToListItem(elementId) {
let element = document.getElementById(elementId);
if (element) {
let parent = element.parentElement;
if (parent) {
parent.scrollTop = element.offsetTop;
}
}
}
//scrolls to the selected element. Default is documentElement (i.e., html element)
scrollTo(selector, left, top, behavior) {
let element = document.querySelector(selector) || document.documentElement;
element.scrollTo({ left, top, behavior });
}
scrollToBottom(selector, behavior) {
let element = document.querySelector(selector);
if (element)
element.scrollTop = element.scrollHeight;
else
window.scrollTo(0, document.body.scrollHeight);
}
//locks the scroll of the selected element. Default is body
lockScroll(selector, lockclass) {
let element = document.querySelector(selector) || document.body;
//if the body doesn't have a scroll bar, don't add the lock class
let hasScrollBar = window.innerWidth > document.body.clientWidth;
if (hasScrollBar) {
element.classList.add(lockclass);
}
}
//unlocks the scroll. Default is body
unlockScroll(selector, lockclass) {
let element = document.querySelector(selector) || document.body;
element.classList.remove(lockclass);
}
};
window.mudScrollManager = new MudScrollManager();
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
//Functions related to the scroll spy
class MudScrollSpy {
constructor() {
this.scrollToSectionRequested = null;
this.lastKnowElement = null;
//needed as variable to remove the event listeners
this.handlerRef = null;
}
// subscribe to relevant events
spying(dotnetReference, selector) {
this.scrollToSectionRequested = null;
this.lastKnowElement = null;
this.handlerRef = this.handleScroll.bind(this, selector, dotnetReference);
// add the event for scroll. In case of zooming this event is also fired
document.addEventListener('scroll', this.handlerRef, true);
// a window resize could change the size of the relevant viewport
window.addEventListener('resize', this.handlerRef, true);
}
// handle the document scroll event and if needed, fires the .NET event
handleScroll(dotnetReference, selector, event) {
const elements = document.getElementsByClassName(selector);
if (elements.length === 0) {
return;
}
const center = window.innerHeight / 2.0;
let minDifference = Number.MAX_SAFE_INTEGER;
let elementId = '';
for (let i = 0; i < elements.length; i++) {
const element = elements[i];
const rect = element.getBoundingClientRect();
const diff = Math.abs(rect.top - center);
if (diff < minDifference) {
minDifference = diff;
elementId = element.id;
}
}
if (document.getElementById(elementId).getBoundingClientRect().top < window.innerHeight * 0.8 === false) {
return;
}
if (this.scrollToSectionRequested != null) {
if (this.scrollToSectionRequested == ' ' && window.scrollY == 0) {
this.scrollToSectionRequested = null;
}
else {
if (elementId === this.scrollToSectionRequested) {
this.scrollToSectionRequested = null;
}
}
return;
}
if (elementId != this.lastKnowElement) {
this.lastKnowElement = elementId;
history.replaceState(null, '', window.location.pathname + "#" + elementId);
dotnetReference.invokeMethodAsync('SectionChangeOccured', elementId);
}
}
activateSection(sectionId) {
const element = document.getElementById(sectionId);
if (element) {
this.lastKnowElement = sectionId;
history.replaceState(null, '', window.location.pathname + "#" + sectionId);
}
}
scrollToSection(sectionId) {
if (sectionId) {
let element = document.getElementById(sectionId);
if (element) {
this.scrollToSectionRequested = sectionId;
element.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'start' });
}
}
else {
window.scrollTo({ top: 0, behavior: 'smooth' });
this.scrollToSectionRequested = ' ';
}
}
//remove event listeners
unspy() {
document.removeEventListener('scroll', this.handlerRef, true);
window.removeEventListener('resize', this.handlerRef, true);
}
};
window.mudScrollSpy = new MudScrollSpy();
const darkThemeMediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
window.darkModeChange = (dotNetHelper) => {
return darkThemeMediaQuery.matches;
};
// Copyright (c) MudBlazor 2021
// MudBlazor licenses this file to you under the MIT license.
// See the LICENSE file in the project root for more information.
class MudWindow {
copyToClipboard (text) {
navigator.clipboard.writeText(text);
}
changeCssById (id, css) {
var element = document.getElementById(id);
if (element) {
element.className = css;
}
}
changeGlobalCssVariable (name, newValue) {
document.documentElement.style.setProperty(name, newValue);
}
// Needed as per https://stackoverflow.com/questions/62769031/how-can-i-open-a-new-window-without-using-js
open (args) {
window.open(args);
}
};
window.mudWindow = new MudWindow();

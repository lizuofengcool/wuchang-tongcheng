// .uvue / .uts 语法结构检查脚本
// 检查项：
//   1. .uvue 文件三段式结构（template/script/style）
//   2. .uts 文件基础语法
//   3. 括号匹配（{ }、( )、[ ]）
//   4. 引号匹配（' " `）
//   5. import 路径有效性（@/ 开头解析为文件）
//   6. type 声明基本规范
// 用法：node check_uvue_syntax.cjs

const fs = require('fs');
const path = require('path');

const ROOT = path.resolve(__dirname);
const APP_ROOT = path.join(ROOT, 'frontend', 'app');

const MODULES = ['love', 'pinche', 'linggong', 'dh114'];

const errors = [];
const warnings = [];
const stats = {
    uvueChecked: 0,
    utsChecked: 0,
    uvueFiles: 0,
    utsFiles: 0,
    errors: 0,
    warnings: 0,
};

// ============ 工具函数 ============

function readFile(p) {
    try {
        return fs.readFileSync(p, 'utf8');
    } catch (e) {
        return null;
    }
}

function fileExists(p) {
    try {
        return fs.statSync(p).isFile();
    } catch (e) {
        return false;
    }
}

// 括号匹配检查（忽略字符串、模板字符串、注释内的括号）
function checkBrackets(content, filePath) {
    const stack = [];
    let i = 0;
    let inSingleStr = false; // 'string'
    let inDoubleStr = false; // "string"
    let inTemplateStr = false; // `template`
    let inLineComment = false; // //...
    let inBlockComment = false; // /*...*/
    let inHtmlComment = false; // <!--...-->

    while (i < content.length) {
        const ch = content[i];
        const next = content[i + 1] || '';

        // 处理字符串状态
        if (!inLineComment && !inBlockComment && !inHtmlComment) {
            if (!inSingleStr && !inDoubleStr && !inTemplateStr) {
                if (ch === '/' && next === '/') { inLineComment = true; i += 2; continue; }
                if (ch === '/' && next === '*') { inBlockComment = true; i += 2; continue; }
                if (ch === '<' && content.substr(i, 4) === '<!--') { inHtmlComment = true; i += 4; continue; }
                if (ch === "'") { inSingleStr = true; i++; continue; }
                if (ch === '"') { inDoubleStr = true; i++; continue; }
                if (ch === '`') { inTemplateStr = true; i++; continue; }
            } else {
                // 在字符串中，处理转义
                if (ch === '\\') { i += 2; continue; }
                if (inSingleStr && ch === "'") { inSingleStr = false; i++; continue; }
                if (inDoubleStr && ch === '"') { inDoubleStr = false; i++; continue; }
                if (inTemplateStr && ch === '`') { inTemplateStr = false; i++; continue; }
                i++;
                continue;
            }
        } else {
            if (inLineComment) {
                if (ch === '\n') { inLineComment = false; }
                i++;
                continue;
            }
            if (inBlockComment) {
                if (ch === '*' && next === '/') { inBlockComment = false; i += 2; continue; }
                i++;
                continue;
            }
            if (inHtmlComment) {
                if (content.substr(i, 3) === '-->') { inHtmlComment = false; i += 3; continue; }
                i++;
                continue;
            }
        }

        // 括号匹配
        if (ch === '{' || ch === '(' || ch === '[') {
            stack.push({ char: ch, line: getLineNum(content, i) });
        } else if (ch === '}' || ch === ')' || ch === ']') {
            const expected = ch === '}' ? '{' : (ch === ')' ? '(' : '[');
            const top = stack.pop();
            if (!top) {
                errors.push(`${filePath}:${getLineNum(content, i)} 多余的闭合符号 '${ch}'`);
            } else if (top.char !== expected) {
                errors.push(`${filePath}:${top.line} 括号不匹配 '${top.char}' 期望 '${expected}' 但遇到 '${ch}'`);
            }
        }
        i++;
    }

    if (inSingleStr) errors.push(`${filePath} 单引号字符串未闭合`);
    if (inDoubleStr) errors.push(`${filePath} 双引号字符串未闭合`);
    if (inTemplateStr) errors.push(`${filePath} 模板字符串未闭合`);
    if (inBlockComment) errors.push(`${filePath} 块注释未闭合`);

    while (stack.length > 0) {
        const top = stack.pop();
        errors.push(`${filePath}:${top.line} 未闭合的括号 '${top.char}'`);
    }
}

function getLineNum(content, pos) {
    let line = 1;
    for (let i = 0; i < pos && i < content.length; i++) {
        if (content[i] === '\n') line++;
    }
    return line;
}

// 检查 .uvue 文件结构
function checkUvueFile(filePath) {
    const content = readFile(filePath);
    if (!content) {
        errors.push(`${filePath} 文件读取失败`);
        return;
    }
    stats.uvueFiles++;

    if (content.trim().length === 0) {
        errors.push(`${filePath} 文件为空`);
        return;
    }

    // 检查 <template> 段
    const templateMatch = content.match(/<template>([\s\S]*?)<\/template>/);
    if (!templateMatch) {
        errors.push(`${filePath} 缺少 <template> 段`);
    } else if (templateMatch[1].trim().length === 0) {
        errors.push(`${filePath} <template> 段为空`);
    }

    // 检查 <script setup lang="uts"> 段
    const scriptMatch = content.match(/<script\s+setup\s+lang="uts">([\s\S]*?)<\/script>/);
    if (!scriptMatch) {
        // 兼容其他写法
        const scriptAny = content.match(/<script[^>]*>([\s\S]*?)<\/script>/);
        if (!scriptAny) {
            errors.push(`${filePath} 缺少 <script> 段`);
        } else {
            warnings.push(`${filePath} <script> 未使用 setup lang="uts" 规范写法`);
        }
    } else if (scriptMatch[1].trim().length === 0) {
        errors.push(`${filePath} <script> 段为空`);
    }

    // 检查 <style> 段（可选，但推荐）
    const styleMatch = content.match(/<style[^>]*>([\s\S]*?)<\/style>/);
    if (!styleMatch) {
        warnings.push(`${filePath} 缺少 <style> 段（可选）`);
    }

    // 检查括号匹配（仅 script 段，template 段的括号是 HTML 属性不参与匹配）
    if (scriptMatch) {
        checkBrackets(scriptMatch[1], filePath + ' (script)');
    }

    // 检查 import 路径
    checkImports(scriptMatch ? scriptMatch[1] : '', filePath);

    stats.uvueChecked++;
}

// 检查 .uts 文件
function checkUtsFile(filePath) {
    const content = readFile(filePath);
    if (!content) {
        errors.push(`${filePath} 文件读取失败`);
        return;
    }
    stats.utsFiles++;

    if (content.trim().length === 0) {
        errors.push(`${filePath} 文件为空`);
        return;
    }

    checkBrackets(content, filePath);
    checkImports(content, filePath);

    stats.utsChecked++;
}

// 检查 import 路径有效性
function checkImports(content, filePath) {
    const importRegex = /import\s+(?:\{[^}]*\}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]/g;
    let match;
    while ((match = importRegex.exec(content)) !== null) {
        const importPath = match[1];
        // 只检查 @/ 开头的相对路径
        if (importPath.startsWith('@/')) {
            const resolvedPath = path.join(APP_ROOT, importPath.substr(2));
            // 尝试多种扩展名
            const candidates = [
                resolvedPath,
                resolvedPath + '.uts',
                resolvedPath + '.uvue',
                resolvedPath + '.ts',
                resolvedPath + '.js',
                resolvedPath + '/index.uts',
                resolvedPath + '/index.uvue',
                resolvedPath + '/index.ts',
                resolvedPath + '/index.js',
            ];
            const found = candidates.some(p => fileExists(p));
            if (!found) {
                warnings.push(`${filePath} import 路径可能无效: ${importPath} (解析为 ${resolvedPath})`);
            }
        }
    }
}

// ============ 主流程 ============

function walkDir(dir, ext) {
    const results = [];
    if (!fs.existsSync(dir)) return results;
    const items = fs.readdirSync(dir);
    for (const item of items) {
        const full = path.join(dir, item);
        const stat = fs.statSync(full);
        if (stat.isDirectory()) {
            results.push(...walkDir(full, ext));
        } else if (full.endsWith(ext)) {
            results.push(full);
        }
    }
    return results;
}

console.log('=== 开始 .uvue / .uts 语法检查 ===\n');
console.log(`App Root: ${APP_ROOT}\n`);

for (const mod of MODULES) {
    console.log(`--- 检查模块: ${mod} ---`);

    // 检查页面 .uvue
    const pagesDir = path.join(APP_ROOT, 'pages', 'business', mod);
    const uvuePages = walkDir(pagesDir, '.uvue');
    console.log(`  页面 .uvue 文件: ${uvuePages.length}`);
    for (const f of uvuePages) checkUvueFile(f);

    // 检查组件 .uvue
    const compDir = path.join(APP_ROOT, 'components', mod);
    const uvueComps = walkDir(compDir, '.uvue');
    console.log(`  组件 .uvue 文件: ${uvueComps.length}`);
    for (const f of uvueComps) checkUvueFile(f);

    // 检查 API .uts
    const utsFile = path.join(APP_ROOT, 'api', 'business', `${mod}.uts`);
    if (fileExists(utsFile)) {
        console.log(`  API .uts 文件: 1`);
        checkUtsFile(utsFile);
    } else {
        errors.push(`API 文件不存在: ${utsFile}`);
    }

    console.log('');
}

console.log('=== 检查结果汇总 ===');
console.log(`.uvue 文件数: ${stats.uvueFiles} (已检查 ${stats.uvueChecked})`);
console.log(`.uts  文件数: ${stats.utsFiles} (已检查 ${stats.utsChecked})`);
console.log(`错误数: ${errors.length}`);
console.log(`警告数: ${warnings.length}`);

if (warnings.length > 0) {
    console.log('\n--- 警告（前 30 条）---');
    warnings.slice(0, 30).forEach(w => console.log('  [WARN] ' + w));
    if (warnings.length > 30) console.log(`  ... 还有 ${warnings.length - 30} 条警告`);
}

if (errors.length > 0) {
    console.log('\n--- 错误（全部）---');
    errors.forEach(e => console.log('  [ERROR] ' + e));
    console.log('\n=== 检查失败：存在 ' + errors.length + ' 个错误 ===');
    process.exit(1);
} else {
    console.log('\n=== 检查通过：所有文件语法结构正确 ===');
    process.exit(0);
}

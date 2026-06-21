package ai

import (
	"bytes"
	"fmt"
	"text/template"
)

// PromptTemplate 是带变量替换的提示词模板。
type PromptTemplate struct {
	Name         string
	SystemPrompt string
	UserTemplate string
}

// Render 使用给定变量渲染用户提示词模板。
func (t *PromptTemplate) Render(vars map[string]string) (string, error) {
	tmpl, err := template.New(t.Name).Parse(t.UserTemplate)
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("execute prompt template: %w", err)
	}
	return buf.String(), nil
}

// ReportSystemPrompt 是报告生成的系统提示词。
const ReportSystemPrompt = `你是一位专业的研究分析师，擅长根据用户给定的标题和主题生成结构清晰、内容详实的研究报告。

要求：
1. 使用 Markdown 格式输出；
2. 报告应包含背景、核心观点、案例分析、结论与建议；
3. 语言为中文，必要时保留专业术语的英文原文；
4. 不要编造无法验证的数据，若引用数据请标注来源或说明为示例。`

// ReportGenerationTemplate 用于生成研究报告。
var ReportGenerationTemplate = PromptTemplate{
	Name:         "report_generation",
	SystemPrompt: ReportSystemPrompt,
	UserTemplate: `请根据以下信息生成一份研究报告：

标题：{{.Title}}
主题：{{.Topic}}

要求：
- 字数控制在 2000 字左右；
- 结构完整，逻辑清晰；
- 直接输出 Markdown 正文，不要添加额外解释。`,
}

// ChatSystemPrompt 是普通对话的系统提示词。
const ChatSystemPrompt = "你是一位有帮助的 AI 助手。请用中文回答用户问题，保持简洁、准确。"

// ChatTemplate 用于普通对话。
var ChatTemplate = PromptTemplate{
	Name:         "chat",
	SystemPrompt: ChatSystemPrompt,
	UserTemplate: "{{.Message}}",
}

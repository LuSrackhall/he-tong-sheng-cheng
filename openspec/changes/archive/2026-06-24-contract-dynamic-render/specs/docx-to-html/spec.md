## ADDED Requirements

### Requirement: docx 文件转换为 HTML

系统 SHALL 提供将 .docx 文件内容转换为 HTML 的能力，用于合同预览功能。

#### Scenario: 基本文档内容转换
- **WHEN** 提供一个包含纯文本段落的 .docx 文件字节数据
- **THEN** 返回有效的 HTML 字符串，每个 `<w:p>` 段落转换为 `<p>` 标签，`<w:r>` + `<w:t>` 中的文本保留为段落内的内联内容

#### Scenario: 粗体和斜体格式保留
- **WHEN** .docx 中包含 `<w:b/>`（粗体）或 `<w:i/>`（斜体）格式的文本
- **THEN** 粗体文本转换为 `<strong>` 标签，斜体文本转换为 `<em>` 标签

#### Scenario: 表格转换
- **WHEN** .docx 中包含 `<w:tbl>` 表格结构
- **THEN** 表格转换为 `<table>` 标签，`<w:tr>` 转换为 `<tr>`，`<w:tc>` 转换为 `<td>`

#### Scenario: 段落对齐方式
- **WHEN** .docx 段落的 `<w:pPr>` 中包含 `<w:jc w:val="center"/>` 或 `<w:jc w:val="right"/>`
- **THEN** 对应的 HTML `<p>` 标签 SHALL 包含 `style="text-align: center"` 或 `style="text-align: right"`

#### Scenario: 段落缩进
- **WHEN** .docx 段落的 `<w:pPr>/<w:ind>` 中包含 `firstLine` 属性
- **THEN** 对应的 HTML `<p>` 标签 SHALL 包含 `style="text-indent: Xem"`，其中 X 由 firstLine 值换算为 em 单位

#### Scenario: 无效或损坏的 docx 文件
- **WHEN** 提供的字节数据不是有效的 ZIP 格式或不包含 `word/document.xml`
- **THEN** 函数 SHALL 返回错误，不产生部分 HTML 输出

#### Scenario: 空文档
- **WHEN** .docx 的 `word/document.xml` 中没有任何段落或表格
- **THEN** 返回空的 HTML 字符串（`""`），不返回错误

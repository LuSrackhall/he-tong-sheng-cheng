## ADDED Requirements

### Requirement: PDF 模板编译缓存
`GenerateReceiptHTML` 和 `GenerateTemplatePreviewHTML` 函数 SHALL 使用 `sync.Once` 在首次调用时解析 HTML 模板字符串，后续调用复用已解析的模板对象，不再重复解析。

#### Scenario: 首次生成收据 HTML
- **WHEN** 第一次调用 `GenerateReceiptHTML`
- **THEN** 解析模板字符串并缓存，返回渲染结果

#### Scenario: 后续生成收据 HTML
- **WHEN** 第 N 次调用 `GenerateReceiptHTML`
- **THEN** 直接使用缓存的模板对象，不重新解析，返回渲染结果

#### Scenario: 模板解析失败
- **WHEN** 模板字符串存在语法错误
- **THEN** `sync.Once` 返回解析错误，后续调用也返回同一错误

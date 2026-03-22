package leetcode

// GraphQL queries for leetcode.com
const (
	queryProblemList = `
query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {
  problemsetQuestionList: questionList(
    categorySlug: $categorySlug
    limit: $limit
    skip: $skip
    filters: $filters
  ) {
    total: totalNum
    questions: data {
      questionId
      questionFrontendId
      title
      titleSlug
      difficulty
      isPaidOnly
      acRate
      topicTags {
        name
        slug
      }
      status
    }
  }
}`

	queryProblemDetail = `
query questionData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    questionFrontendId
    title
    titleSlug
    difficulty
    content
    isPaidOnly
    topicTags {
      name
      slug
    }
    codeSnippets {
      lang
      langSlug
      code
    }
    sampleTestCase
  }
}`

	queryRunCode = `
mutation runCode($titleSlug: String!, $code: String!, $lang: String!, $dataInput: String!) {
  runCode(
    titleSlug: $titleSlug
    code: $code
    lang: $lang
    dataInput: $dataInput
  ) {
    interpret_id
  }
}`

	querySubmitCode = `
mutation submitCode($titleSlug: String!, $code: String!, $lang: String!, $questionId: Int!) {
  submitCode(
    titleSlug: $titleSlug
    code: $code
    lang: $lang
    questionId: $questionId
  ) {
    submission_id
  }
}`

	queryCheckSubmission = `
query checkSubmission($id: Int!) {
  submissionDetails(submissionId: $id) {
    statusCode
    statusDisplay
    runtimeDisplay
    memoryDisplay
    totalCorrect
    totalTestcases
  }
}`

	querySubmissionList = `
query submissionList($questionSlug: String!, $offset: Int!, $limit: Int!) {
  questionSubmissionList(questionSlug: $questionSlug, offset: $offset, limit: $limit) {
    submissions {
      id
      lang
      statusDisplay
      timestamp
    }
  }
}`

	querySubmissionDetail = `
query submissionDetails($submissionId: Int!) {
  submissionDetails(submissionId: $submissionId) {
    code
    statusDisplay
    timestamp
  }
}`
)

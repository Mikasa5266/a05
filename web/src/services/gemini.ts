// Mock implementation of Gemini service
// In a real application, this would call the Google GenAI API or a backend proxy

export const parseResume = async (file) => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 2000))

  // Mock data based on file content (simulated)
  return {
    techStack: ['Vue 3', 'TypeScript', 'Tailwind CSS', 'Node.js', 'Go'],
    intent: '高级前端工程师 / 全栈工程师',
    softSkills: ['团队协作', '问题解决', '敏捷开发', '快速学习'],
    experience: [
      {
        title: 'Senior Frontend Developer',
        description: 'Led the frontend team in rebuilding the legacy application using Vue 3.',
        highlights: ['Improved performance by 40%', 'Implemented CI/CD pipeline']
      }
    ]
  }
}

export const generateQuestions = async (resumeData, mode) => {
  await new Promise(resolve => setTimeout(resolve, 1500))

  return [
    { 
      id: '1',
      text: '请介绍一下你最深入参与的一个项目，以及你在其中解决的最大的技术难点是什么？', 
      type: 'technical',
      context: '考察技术深度和解决问题的能力' 
    },
    { 
      id: '2',
      text: '在 Vue 3 中，Composition API 与 Options API 相比有哪些优势？', 
      type: 'technical',
      context: '考察框架理解' 
    },
    { 
      id: '3',
      text: '如何设计一个高可用的微服务架构？', 
      type: 'technical',
      context: '考察系统设计能力' 
    },
    { 
      id: '4',
      text: '你如何处理团队中的技术分歧？', 
      type: 'behavioral',
      context: '考察软技能' 
    },
    { 
      id: '5',
      text: '你对未来的职业规划是什么？', 
      type: 'behavioral',
      context: '考察稳定性与成长性' 
    }
  ]
}

export const evaluateAnswer = async (question, answer) => {
  await new Promise(resolve => setTimeout(resolve, 2000))

  return {
    score: 85,
    feedback: '回答逻辑清晰，涵盖了主要的技术点。但在具体案例的描述上可以更详细一些。',
    metrics: {
      technical: 85,
      expression: 90,
      logic: 80
    }
  }
}

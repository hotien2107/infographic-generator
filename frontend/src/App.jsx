import { useProjectWorkflow } from '@/hooks/useProjectWorkflow'

import { ApiChecklistCard } from '@/components/workflow/ApiChecklistCard'
import { DocumentList } from '@/components/workflow/DocumentList'
import { FeedbackBanner } from '@/components/workflow/FeedbackBanner'
import { HeroCard } from '@/components/workflow/HeroCard'
import { ProcessingSummaryCard } from '@/components/workflow/ProcessingSummaryCard'
import { ProjectSnapshot } from '@/components/workflow/ProjectSnapshot'

function App() {
  const { state, actions } = useProjectWorkflow()

  return (
    <main className="min-h-screen px-4 py-10 text-foreground sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-7xl flex-col gap-6">
        <section className="grid gap-6 lg:grid-cols-[1.35fr_0.65fr]">
          <HeroCard
            title={state.title}
            inputMode={state.inputMode}
            notes={state.notes}
            isCreating={state.isCreating}
            projectId={state.projectId}
            canUpload={state.canUpload}
            canTriggerProcessing={state.canTriggerProcessing}
            isLoading={state.isLoading}
            isUploading={state.isUploading}
            isTriggering={state.isTriggering}
            onTitleChange={actions.setTitle}
            onInputModeChange={actions.setInputMode}
            onNotesChange={actions.setNotes}
            onSubmit={actions.handleCreateProject}
            onProjectIdChange={actions.setProjectId}
            onLoadProject={actions.handleLoadProject}
            onFileChange={actions.setSelectedFile}
            onUpload={actions.handleUpload}
            onTriggerProcessing={actions.handleTriggerProcessing}
          />
          <ProjectSnapshot project={state.project} />
        </section>

        <FeedbackBanner errorMessage={state.errorMessage} successMessage={state.successMessage} />

        <section className="grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
          <ApiChecklistCard />
          <ProcessingSummaryCard project={state.project} />
        </section>

        <DocumentList documents={state.project?.documents ?? []} />
      </div>
    </main>
  )
}

export default App

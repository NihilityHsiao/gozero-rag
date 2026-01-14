import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from '@/components/layout/MainLayout';
import KnowledgeList from '@/pages/knowledge/KnowledgeList';
import CreateKnowledge from '@/pages/knowledge/CreateKnowledge';
import Login from '@/pages/auth/Login';
import Register from '@/pages/auth/Register';
import ProtectedRoute from '@/components/layout/ProtectedRoute';
import { Toaster } from 'sonner';

import KnowledgeDetailLayout from '@/components/layout/KnowledgeDetailLayout';
import DocumentList from '@/pages/knowledge/DocumentList';
import DatasetCreatePage from '@/pages/dataset/DatasetCreatePage';
import DocumentEditPage from '@/pages/dataset/DocumentEditPage';
import KnowledgeSettings from '@/pages/knowledge/KnowledgeSettings';
import { SettingsLayout, ModelList, UserInfo, TeamManagement } from '@/pages/settings';

import DocumentDetailPage from '@/pages/knowledge/document/DocumentDetailPage';
import RetrievalTestPage from '@/pages/knowledge/RetrievalTestPage';
import ChatPage from '@/pages/chat';

function App() {
  return (
    <BrowserRouter>
      <Toaster richColors position="top-center" />
      <Routes>
        {/* 认证路由 */}
        <Route path="/login" element={<Login />} />
        <Route path="/auth/login" element={<Login />} />
        <Route path="/auth/register" element={<Register />} />

        {/* Protected Routes */}
        <Route element={<ProtectedRoute />}>
          <Route path="/" element={<MainLayout />}>
            <Route index element={<Navigate to="/knowledge" replace />} />
            <Route path="knowledge" element={<KnowledgeList />} />
            <Route path="knowledge/create" element={<CreateKnowledge />} />
            <Route path="chat" element={<ChatPage />} />
            <Route path="chat/:conversationId" element={<ChatPage />} />
          </Route>

          {/* Dataset Creation Wizard (Standalone) */}
          <Route path="/knowledge/:id/dataset/create" element={<DatasetCreatePage />} />

          {/* Document Edit Page (Standalone) */}
          <Route path="/knowledge/:id/document/:doc_id/edit" element={<DocumentEditPage />} />

          {/* Knowledge Detail Routes */}
          <Route path="knowledge/:id" element={<KnowledgeDetailLayout />}>
            <Route index element={<Navigate to="documents" replace />} />
            <Route path="documents" element={<DocumentList />} />
            <Route path="document/:docId" element={<DocumentDetailPage />} />
            <Route path="settings" element={<KnowledgeSettings />} />
            <Route path="retrieve" element={<RetrievalTestPage />} />
          </Route>

          {/* User Settings Routes */}
          <Route path="settings" element={<SettingsLayout />}>
            <Route index element={<Navigate to="provider" replace />} />
            <Route path="provider" element={<ModelList />} />
            <Route path="profile" element={<UserInfo />} />
            <Route path="team" element={<TeamManagement />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

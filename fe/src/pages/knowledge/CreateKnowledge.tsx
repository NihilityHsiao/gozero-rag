import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { toast } from 'sonner';
import { createKnowledgeBase } from '@/api/knowledge';
import { getUserApiList } from '@/api/user_model';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { FileText, Globe, Cpu } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/store/useAuthStore';
import type { UserApiInfo } from '@/types';

// Zod schema for form validation
const formSchema = z.object({
  name: z.string().min(1, { message: '请输入知识库名称' }),
  description: z.string().optional(),
  embedding_id: z.coerce.number().min(1, { message: '请选择 Embedding 模型' }),
  rerank_id: z.coerce.number().optional(),
  rewrite_id: z.coerce.number().optional(),
  qa_id: z.coerce.number().optional(),
  chat_id: z.coerce.number().optional(),
});

type FormValues = z.infer<typeof formSchema>;

export default function CreateKnowledge() {
  const navigate = useNavigate();
  const { userInfo } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [dataSourceType, setDataSourceType] = useState<'file' | 'web'>('file');
  const [models, setModels] = useState<UserApiInfo[]>([]);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema) as any,
    defaultValues: {
      name: '',
      description: '',
      embedding_id: 0,
      rerank_id: 0,
      rewrite_id: 0,
      qa_id: 0,
      chat_id: 0,
    },
    mode: 'onChange',
  });

  useEffect(() => {
    if (userInfo?.user_id) {
      getUserApiList(userInfo.user_id, { status: 1 }).then((res) => {
        // @ts-ignore
        const list = res.list || res.data?.list || [];
        setModels(list);

        // Auto-select defaults
        const defaultEmbedding = list.find((m: UserApiInfo) => m.model_type === 'embedding' && m.is_default === 1);
        if (defaultEmbedding) form.setValue('embedding_id', defaultEmbedding.id);

        const defaultRerank = list.find((m: UserApiInfo) => m.model_type === 'rerank' && m.is_default === 1);
        if (defaultRerank) form.setValue('rerank_id', defaultRerank.id);

        const defaultRewrite = list.find((m: UserApiInfo) => m.model_type === 'rewrite' && m.is_default === 1);
        if (defaultRewrite) form.setValue('rewrite_id', defaultRewrite.id);

        const defaultQa = list.find((m: UserApiInfo) => m.model_type === 'qa' && m.is_default === 1);
        if (defaultQa) form.setValue('qa_id', defaultQa.id);

        const defaultChat = list.find((m: UserApiInfo) => m.model_type === 'chat' && m.is_default === 1);
        if (defaultChat) form.setValue('chat_id', defaultChat.id);

      }).catch(err => {
        console.error("Failed to fetch models", err);
        toast.error("加载模型失败");
      });
    }
  }, [userInfo, form]);

  const getModelsByType = (type: string) => models.filter(m => m.model_type === type);

  const onSubmit = async (values: FormValues) => {
    setLoading(true);
    try {
      const res = await createKnowledgeBase(values);
      if (res) {
        toast.success('知识库创建成功');
        navigate('/knowledge');
      }
    } catch (error) {
      console.error('Failed to create knowledge base:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col h-full bg-gray-50">
      {/* Header */}
      <header className="flex-shrink-0 h-16 bg-white border-b border-gray-200 px-6 lg:px-8 flex items-center justify-between sticky top-0 z-10">
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span className="font-semibold text-gray-900 text-lg">创建知识库</span>
        </div>
      </header>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6 lg:p-8">
        <div className="max-w-3xl mx-auto space-y-6">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit as any)} className="space-y-8">

              {/* Basic Information Section */}
              <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                <h2 className="text-lg font-medium text-gray-900 mb-4">基本信息</h2>
                <div className="space-y-4">
                  <FormField
                    control={form.control as any}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>名称</FormLabel>
                        <FormControl>
                          <Input placeholder="输入知识库名称" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control as any}
                    name="description"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>描述</FormLabel>
                        <FormControl>
                          <Textarea
                            placeholder="描述知识库的用途"
                            className="resize-none min-h-[100px]"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              {/* Model Configuration Section */}
              <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                <div className="flex items-center gap-2 mb-4">
                  <Cpu className="w-5 h-5 text-gray-500" />
                  <h2 className="text-lg font-medium text-gray-900">模型配置</h2>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Embedding Model (Required) */}
                  <FormField
                    control={form.control as any}
                    name="embedding_id"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Embedding 模型 <span className="text-red-500">*</span></FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          defaultValue={field.value ? String(field.value) : undefined}
                          value={field.value ? String(field.value) : undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择 Embedding 模型" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {getModelsByType('embedding').map((model) => (
                              <SelectItem key={model.id} value={String(model.id)}>
                                {model.config_name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Rerank Model */}
                  <FormField
                    control={form.control as any}
                    name="rerank_id"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Rerank 模型</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          defaultValue={field.value ? String(field.value) : undefined}
                          value={field.value ? String(field.value) : undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择 Rerank 模型 (可选)" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="0">无</SelectItem>
                            {getModelsByType('rerank').map((model) => (
                              <SelectItem key={model.id} value={String(model.id)}>
                                {model.config_name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Rewrite Model */}
                  <FormField
                    control={form.control as any}
                    name="rewrite_id"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Rewrite 模型</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          defaultValue={field.value ? String(field.value) : undefined}
                          value={field.value ? String(field.value) : undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择 Rewrite 模型 (可选)" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="0">无</SelectItem>
                            {getModelsByType('rewrite').map((model) => (
                              <SelectItem key={model.id} value={String(model.id)}>
                                {model.config_name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* QA Model */}
                  <FormField
                    control={form.control as any}
                    name="qa_id"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>QA 模型</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          defaultValue={field.value ? String(field.value) : undefined}
                          value={field.value ? String(field.value) : undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择 QA 模型 (可选)" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="0">无</SelectItem>
                            {getModelsByType('qa').map((model) => (
                              <SelectItem key={model.id} value={String(model.id)}>
                                {model.config_name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  {/* Chat Model */}
                  <FormField
                    control={form.control as any}
                    name="chat_id"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>对话模型</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          defaultValue={field.value ? String(field.value) : undefined}
                          value={field.value ? String(field.value) : undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择对话模型 (可选)" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="0">无</SelectItem>
                            {getModelsByType('chat').map((model) => (
                              <SelectItem key={model.id} value={String(model.id)}>
                                {model.config_name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              {/* Data Source Section (Mock UI) */}
              <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                <h2 className="text-lg font-medium text-gray-900 mb-4">数据源</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Import from Text/File Card */}
                  <div
                    className={cn(
                      "cursor-pointer rounded-xl border-2 p-4 transition-all hover:border-blue-500 hover:bg-blue-50/50",
                      dataSourceType === 'file' ? "border-blue-600 bg-blue-50" : "border-gray-200 bg-white"
                    )}
                    onClick={() => setDataSourceType('file')}
                  >
                    <div className="flex items-start gap-3">
                      <div className={cn(
                        "p-2 rounded-lg",
                        dataSourceType === 'file' ? "bg-blue-100 text-blue-600" : "bg-gray-100 text-gray-500"
                      )}>
                        <FileText size={24} />
                      </div>
                      <div>
                        <h3 className="font-medium text-gray-900">导入文本 / 文件</h3>
                        <p className="text-sm text-gray-500 mt-1">支持 PDF, Word, Markdown 等</p>
                      </div>
                    </div>
                  </div>

                  {/* Sync from Website Card */}
                  <div
                    className={cn(
                      "cursor-pointer rounded-xl border-2 p-4 transition-all hover:border-blue-500 hover:bg-blue-50/50",
                      dataSourceType === 'web' ? "border-blue-600 bg-blue-50" : "border-gray-200 bg-white"
                    )}
                    onClick={() => setDataSourceType('web')}
                  >
                    <div className="flex items-start gap-3">
                      <div className={cn(
                        "p-2 rounded-lg",
                        dataSourceType === 'web' ? "bg-blue-100 text-blue-600" : "bg-gray-100 text-gray-500"
                      )}>
                        <Globe size={24} />
                      </div>
                      <div>
                        <h3 className="font-medium text-gray-900">同步站点</h3>
                        <p className="text-sm text-gray-500 mt-1">通过 Sitemap 或 URL 同步</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate('/knowledge')}
                  disabled={loading}
                >
                  取消
                </Button>
                <Button type="submit" disabled={loading}>
                  {loading ? '创建中...' : '确认创建'}
                </Button>
              </div>

            </form>
          </Form>
        </div>
      </div>
    </div>
  );
}

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { toast } from 'sonner';
import { createKnowledgeBase, getTenantLlmList } from '@/api/knowledge';
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
import { Cpu, Loader2 } from 'lucide-react';

// Zod schema for form validation
const formSchema = z.object({
  name: z.string().min(1, { message: '请输入知识库名称' }),
  description: z.string().optional(),
  embd_id: z.string().min(1, { message: '请选择 Embedding 模型' }),
});

type FormValues = z.infer<typeof formSchema>;

interface TenantLlmModel {
  id: number;
  llm_name: string;
  llm_factory: string;
  model_type: string;
}

export default function CreateKnowledge() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [embeddingModels, setEmbeddingModels] = useState<TenantLlmModel[]>([]);
  const [loadingModels, setLoadingModels] = useState(true);

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      embd_id: '',
    },
    mode: 'onChange',
  });

  // 加载租户 Embedding 模型
  useEffect(() => {
    const fetchModels = async () => {
      setLoadingModels(true);
      try {
        const res = await getTenantLlmList({ model_type: 'embedding' });
        const models = res.list || [];
        setEmbeddingModels(models);

        // 自动选择第一个模型
        if (models.length > 0) {
          const firstModel = models[0];
          form.setValue('embd_id', `${firstModel.llm_name}@${firstModel.llm_factory}`);
        }
      } catch (err) {
        console.error("Failed to fetch tenant LLM models", err);
        toast.error("加载租户模型失败");
      } finally {
        setLoadingModels(false);
      }
    };
    fetchModels();
  }, [form]);

  const onSubmit = async (values: FormValues) => {
    setLoading(true);
    try {
      await createKnowledgeBase(values);
      toast.success('知识库创建成功');
      navigate('/knowledge');
    } catch (error: any) {
      console.error('Failed to create knowledge base:', error);
      toast.error(error?.msg || '创建知识库失败');
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
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">

              {/* Basic Information Section */}
              <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
                <h2 className="text-lg font-medium text-gray-900 mb-4">基本信息</h2>
                <div className="space-y-4">
                  <FormField
                    control={form.control}
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
                    control={form.control}
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
                  <h2 className="text-lg font-medium text-gray-900">Embedding 模型</h2>
                </div>

                <FormField
                  control={form.control}
                  name="embd_id"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>选择模型 <span className="text-red-500">*</span></FormLabel>
                      {loadingModels ? (
                        <div className="flex items-center gap-2 py-2 text-sm text-muted-foreground">
                          <Loader2 className="w-4 h-4 animate-spin" />
                          加载模型中...
                        </div>
                      ) : embeddingModels.length === 0 ? (
                        <div className="py-3 px-4 bg-yellow-50 border border-yellow-200 rounded-lg text-sm text-yellow-800">
                          当前租户未配置 Embedding 模型，请先在"设置 → 模型配置"中添加。
                        </div>
                      ) : (
                        <Select
                          onValueChange={field.onChange}
                          value={field.value}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="选择 Embedding 模型" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {embeddingModels.map((model) => (
                              <SelectItem
                                key={model.id}
                                value={`${model.llm_name}@${model.llm_factory}`}
                              >
                                {model.llm_name} ({model.llm_factory})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                      <FormMessage />
                    </FormItem>
                  )}
                />
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
                <Button
                  type="submit"
                  disabled={loading || loadingModels || embeddingModels.length === 0}
                >
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

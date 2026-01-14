import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useNavigate, Link } from 'react-router-dom';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';

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
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { login } from '@/api/auth';
import { useAuthStore } from '@/store/useAuthStore';

// 使用 email 登录的表单验证
const formSchema = z.object({
  email: z.string().email({ message: '请输入有效的邮箱地址' }),
  password: z.string().min(1, { message: '请输入密码' }),
});

export default function Login() {
  const navigate = useNavigate();
  const { login: setAuth } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);

  // 初始化表单
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  // 提交处理
  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsLoading(true);
    try {
      const res = await login(values);
      // 登录成功
      setAuth(res);
      toast.success('登录成功');
      navigate('/knowledge', { replace: true });
    } catch (error: any) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-[400px] shadow-sm border-gray-200">
        <CardHeader className="space-y-1 pb-6">
          <CardTitle className="text-2xl font-semibold text-center text-gray-900">
            欢迎回来
          </CardTitle>
          <CardDescription className="text-center text-gray-500">
            请输入您的邮箱和密码登录
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-sm font-medium">邮箱</FormLabel>
                    <FormControl>
                      <Input
                        type="email"
                        placeholder="your@email.com"
                        {...field}
                        className="h-10 focus-visible:ring-blue-600"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel className="text-sm font-medium">密码</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        placeholder="••••••••"
                        {...field}
                        className="h-10 focus-visible:ring-blue-600"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button
                type="submit"
                className="w-full bg-blue-600 hover:bg-blue-700 font-bold mt-2"
                disabled={isLoading}
              >
                {isLoading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    登录中...
                  </>
                ) : (
                  '登录'
                )}
              </Button>
            </form>
          </Form>
          <div className="mt-4 text-center text-sm text-gray-500">
            还没有账号？{' '}
            <Link to="/auth/register" className="text-blue-600 hover:underline">
              立即注册
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

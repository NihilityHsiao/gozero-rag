import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/store/useAuthStore";

export default function UserInfo() {
    const { userInfo } = useAuthStore();

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle>个人信息</CardTitle>
                    <CardDescription>
                        查看您的基本账户信息。
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid gap-2">
                        <Label>用户名</Label>
                        <Input value={userInfo?.nickname || ''} disabled readOnly />
                    </div>
                    {/* Add more fields here */}
                </CardContent>
            </Card>
        </div>
    );
}

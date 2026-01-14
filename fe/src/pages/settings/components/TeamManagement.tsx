import { useState, useEffect } from 'react';
import { Users, UserPlus, LogIn, Copy, Check, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { useAuthStore } from '@/store/useAuthStore';
import { listMembers, listJoinedTeams, createInvite, verifyInvite, joinTeam } from '@/api/team';
import type { TeamMember, JoinedTeam, VerifyInviteResp } from '@/api/team';


// 角色显示映射
const roleMap: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' }> = {
    owner: { label: '所有者', variant: 'default' },
    admin: { label: '管理员', variant: 'secondary' },
    member: { label: '成员', variant: 'outline' },
};

// 格式化时间戳
const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleDateString('zh-CN');
};

export default function TeamManagement() {
    const { currentTenant, userInfo } = useAuthStore();
    const [members, setMembers] = useState<TeamMember[]>([]);
    const [joinedTeams, setJoinedTeams] = useState<JoinedTeam[]>([]);
    const [loading, setLoading] = useState(true);

    // 邀请弹窗状态
    const [inviteOpen, setInviteOpen] = useState(false);
    const [inviteEmail, setInviteEmail] = useState('');
    const [inviteLoading, setInviteLoading] = useState(false);
    const [inviteResult, setInviteResult] = useState<{ code: string; link: string } | null>(null);
    const [copied, setCopied] = useState(false);

    // 加入团队弹窗状态
    const [joinOpen, setJoinOpen] = useState(false);
    const [joinCode, setJoinCode] = useState('');
    const [joinLoading, setJoinLoading] = useState(false);
    const [verifyResult, setVerifyResult] = useState<VerifyInviteResp | null>(null);
    const [confirmStep, setConfirmStep] = useState(false);

    // 获取当前用户角色
    const currentRole = members.find(m => m.user_id === userInfo?.user_id)?.role || 'member';
    const canInvite = currentRole === 'owner' || currentRole === 'admin';

    // 加载数据
    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            try {
                const [membersRes, teamsRes] = await Promise.all([
                    listMembers(),
                    listJoinedTeams(),
                ]);
                setMembers(membersRes.list || []);
                setJoinedTeams(teamsRes.list || []);
            } catch (error) {
                console.error('加载团队数据失败:', error);
                toast.error('加载团队数据失败');
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, []);

    // 发起邀请
    const handleInvite = async () => {
        if (!inviteEmail.trim()) {
            toast.error('请输入邮箱');
            return;
        }
        setInviteLoading(true);
        try {
            const res = await createInvite(inviteEmail.trim());
            setInviteResult({ code: res.invite_code, link: res.invite_link });
            toast.success('邀请码已生成');
        } catch (error: any) {
            toast.error(error?.msg || '邀请失败');
        } finally {
            setInviteLoading(false);
        }
    };

    // 复制邀请链接
    const handleCopy = async () => {
        if (inviteResult) {
            await navigator.clipboard.writeText(inviteResult.link);
            setCopied(true);
            toast.success('已复制到剪贴板');
            setTimeout(() => setCopied(false), 2000);
        }
    };

    // 验证邀请码
    const handleVerify = async () => {
        if (!joinCode.trim()) {
            toast.error('请输入邀请码');
            return;
        }
        setJoinLoading(true);
        try {
            const res = await verifyInvite(joinCode.trim());
            setVerifyResult(res);
            setConfirmStep(true);
        } catch (error: any) {
            toast.error(error?.msg || '邀请码无效');
        } finally {
            setJoinLoading(false);
        }
    };

    // 确认加入团队
    const handleJoin = async () => {
        setJoinLoading(true);
        try {
            await joinTeam(joinCode.trim());
            toast.success('加入成功！');
            setJoinOpen(false);
            setJoinCode('');
            setVerifyResult(null);
            setConfirmStep(false);
            // 刷新团队列表
            const teamsRes = await listJoinedTeams();
            setJoinedTeams(teamsRes.list || []);
        } catch (error: any) {
            toast.error(error?.msg || '加入失败');
        } finally {
            setJoinLoading(false);
        }
    };

    // 重置邀请弹窗
    const resetInviteDialog = () => {
        setInviteEmail('');
        setInviteResult(null);
        setCopied(false);
    };

    // 重置加入弹窗
    const resetJoinDialog = () => {
        setJoinCode('');
        setVerifyResult(null);
        setConfirmStep(false);
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 头部: 工作空间名称和操作按钮 */}
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-lg font-medium">{currentTenant?.name || '我的工作空间'}</h3>
                    <p className="text-sm text-muted-foreground">管理团队成员和加入的团队</p>
                </div>
                <div className="flex gap-2">
                    {/* 邀请成员按钮 */}
                    {canInvite && (
                        <Dialog open={inviteOpen} onOpenChange={(open) => { setInviteOpen(open); if (!open) resetInviteDialog(); }}>
                            <DialogTrigger asChild>
                                <Button variant="default" size="sm">
                                    <UserPlus className="w-4 h-4 mr-2" />
                                    邀请成员
                                </Button>
                            </DialogTrigger>
                            <DialogContent>
                                <DialogHeader>
                                    <DialogTitle>邀请成员加入团队</DialogTitle>
                                    <DialogDescription>输入对方的注册邮箱，生成邀请链接</DialogDescription>
                                </DialogHeader>
                                {!inviteResult ? (
                                    <div className="space-y-4 py-4">
                                        <Input
                                            placeholder="请输入邮箱地址"
                                            type="email"
                                            value={inviteEmail}
                                            onChange={(e) => setInviteEmail(e.target.value)}
                                        />
                                    </div>
                                ) : (
                                    <div className="space-y-4 py-4">
                                        <div className="p-4 bg-muted rounded-lg">
                                            <p className="text-sm text-muted-foreground mb-2">邀请码</p>
                                            <p className="font-mono text-lg">{inviteResult.code}</p>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <Input value={inviteResult.link} readOnly className="flex-1" />
                                            <Button variant="outline" size="icon" onClick={handleCopy}>
                                                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                                            </Button>
                                        </div>
                                        <p className="text-xs text-muted-foreground">邀请链接 24 小时内有效，仅限使用一次</p>
                                    </div>
                                )}
                                <DialogFooter>
                                    {!inviteResult ? (
                                        <Button onClick={handleInvite} disabled={inviteLoading}>
                                            {inviteLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                                            生成邀请码
                                        </Button>
                                    ) : (
                                        <Button variant="outline" onClick={() => setInviteOpen(false)}>
                                            完成
                                        </Button>
                                    )}
                                </DialogFooter>
                            </DialogContent>
                        </Dialog>
                    )}

                    {/* 加入团队按钮 */}
                    <Dialog open={joinOpen} onOpenChange={(open) => { setJoinOpen(open); if (!open) resetJoinDialog(); }}>
                        <DialogTrigger asChild>
                            <Button variant="outline" size="sm">
                                <LogIn className="w-4 h-4 mr-2" />
                                加入团队
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>{confirmStep ? '确认加入团队' : '加入团队'}</DialogTitle>
                                <DialogDescription>
                                    {confirmStep ? '请确认以下团队信息' : '输入邀请码加入其他团队'}
                                </DialogDescription>
                            </DialogHeader>
                            {!confirmStep ? (
                                <div className="space-y-4 py-4">
                                    <Input
                                        placeholder="请输入邀请码"
                                        value={joinCode}
                                        onChange={(e) => setJoinCode(e.target.value)}
                                    />
                                </div>
                            ) : (
                                <div className="space-y-4 py-4">
                                    <Card>
                                        <CardContent className="pt-6">
                                            <div className="space-y-2">
                                                <div className="flex justify-between">
                                                    <span className="text-muted-foreground">团队名称</span>
                                                    <span className="font-medium">{verifyResult?.tenant_name}</span>
                                                </div>
                                                <div className="flex justify-between">
                                                    <span className="text-muted-foreground">邀请人</span>
                                                    <span>{verifyResult?.inviter}</span>
                                                </div>
                                            </div>
                                        </CardContent>
                                    </Card>
                                </div>
                            )}
                            <DialogFooter>
                                {!confirmStep ? (
                                    <Button onClick={handleVerify} disabled={joinLoading}>
                                        {joinLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                                        下一步
                                    </Button>
                                ) : (
                                    <div className="flex gap-2">
                                        <Button variant="outline" onClick={() => setConfirmStep(false)}>
                                            返回
                                        </Button>
                                        <Button onClick={handleJoin} disabled={joinLoading}>
                                            {joinLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                                            确认加入
                                        </Button>
                                    </div>
                                )}
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            {/* 团队成员列表 */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Users className="w-5 h-5" />
                        团队成员
                    </CardTitle>
                    <CardDescription>当前工作空间的所有成员</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>姓名</TableHead>
                                <TableHead>邮箱</TableHead>
                                <TableHead>角色</TableHead>
                                <TableHead>加入日期</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {members.map((member) => (
                                <TableRow key={member.user_id}>
                                    <TableCell className="font-medium">{member.nickname}</TableCell>
                                    <TableCell>{member.email}</TableCell>
                                    <TableCell>
                                        <Badge variant={roleMap[member.role]?.variant || 'outline'}>
                                            {roleMap[member.role]?.label || member.role}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>{formatTime(member.joined_time)}</TableCell>
                                </TableRow>
                            ))}
                            {members.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={4} className="text-center text-muted-foreground">
                                        暂无成员
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            {/* 加入的团队列表 */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Users className="w-5 h-5" />
                        加入的团队
                    </CardTitle>
                    <CardDescription>您加入的所有团队工作空间</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>团队名称</TableHead>
                                <TableHead>所有者</TableHead>
                                <TableHead>角色</TableHead>
                                <TableHead>加入日期</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {joinedTeams.map((team) => (
                                <TableRow key={team.tenant_id}>
                                    <TableCell className="font-medium">{team.tenant_name}</TableCell>
                                    <TableCell>{team.owner_name}</TableCell>
                                    <TableCell>
                                        <Badge variant={roleMap[team.role]?.variant || 'outline'}>
                                            {roleMap[team.role]?.label || team.role}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>{formatTime(team.joined_time)}</TableCell>
                                </TableRow>
                            ))}
                            {joinedTeams.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={4} className="text-center text-muted-foreground">
                                        暂无加入的团队
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}

import { Search, Filter } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

interface FloatingSearchProps {
    searchQuery: string;
    onSearchChange: (value: string) => void;
    onSearchEnter: () => void;
}

export function FloatingSearch({ searchQuery, onSearchChange, onSearchEnter }: FloatingSearchProps) {
    return (
        <div className="absolute top-6 left-6 z-20 w-80">
            <div className="bg-slate-900/80 backdrop-blur-md border border-slate-700/50 rounded-xl shadow-xl p-1 flex items-center group ring-1 ring-white/5 transition-all focus-within:ring-indigo-500/50 focus-within:border-indigo-500/50">
                <div className="pl-3 text-slate-400 group-focus-within:text-indigo-400 transition-colors">
                    <Search size={18} />
                </div>
                <Input
                    className="bg-transparent border-none text-white placeholder-slate-500 focus-visible:ring-0 h-10 text-sm"
                    placeholder="Search nodes..."
                    value={searchQuery}
                    onChange={(e) => onSearchChange(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && onSearchEnter()}
                />
                <div className="flex items-center gap-1 pr-1">
                    <Button
                        size="icon"
                        variant="ghost"
                        className="h-8 w-8 text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-lg"
                        title="Filter"
                    >
                        <Filter size={16} />
                    </Button>
                </div>
            </div>

            {/* Quick Filters - Mock Data for now */}
            <div className="mt-3 flex gap-2">
                <Badge variant="outline" className="cursor-pointer bg-slate-900/40 border-slate-700 hover:border-indigo-500/50 hover:bg-indigo-500/10 text-slate-300 hover:text-indigo-300 transition-all">
                    Person
                </Badge>
                <Badge variant="outline" className="cursor-pointer bg-slate-900/40 border-slate-700 hover:border-purple-500/50 hover:bg-purple-500/10 text-slate-300 hover:text-purple-300 transition-all">
                    Organization
                </Badge>
                <Badge variant="outline" className="cursor-pointer bg-slate-900/40 border-slate-700 hover:border-emerald-500/50 hover:bg-emerald-500/10 text-slate-300 hover:text-emerald-300 transition-all">
                    Event
                </Badge>
            </div>
        </div>
    );
}

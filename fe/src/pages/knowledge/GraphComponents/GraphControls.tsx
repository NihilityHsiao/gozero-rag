import { ZoomIn, ZoomOut, Maximize, Focus } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface GraphControlsProps {
    onZoomIn: () => void;
    onZoomOut: () => void;
    onZoomToFit: () => void;
    onFocus?: () => void;
}

export function GraphControls({ onZoomIn, onZoomOut, onZoomToFit, onFocus }: GraphControlsProps) {
    return (
        <div className="absolute bottom-6 right-6 z-20 flex flex-col gap-2">
            <div className="bg-slate-900/80 backdrop-blur-md border border-slate-700/50 rounded-xl shadow-lg p-1 flex flex-col gap-1 ring-1 ring-white/5">
                <Button
                    size="icon"
                    variant="ghost"
                    className="h-10 w-10 text-slate-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                    onClick={onZoomIn}
                    title="Zoom In"
                >
                    <ZoomIn size={20} />
                </Button>
                <Button
                    size="icon"
                    variant="ghost"
                    className="h-10 w-10 text-slate-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors"
                    onClick={onZoomOut}
                    title="Zoom Out"
                >
                    <ZoomOut size={20} />
                </Button>
            </div>

            <Button
                size="icon"
                variant="ghost"
                className="h-10 w-10 bg-slate-900/80 backdrop-blur-md border border-slate-700/50 shadow-lg text-slate-400 hover:text-white hover:bg-white/10 rounded-xl transition-colors ring-1 ring-white/5"
                onClick={onZoomToFit}
                title="Fit to Screen"
            >
                <Maximize size={20} />
            </Button>

            {onFocus && (
                <Button
                    size="icon"
                    variant="ghost"
                    className="h-10 w-10 bg-slate-900/80 backdrop-blur-md border border-slate-700/50 shadow-lg text-slate-400 hover:text-white hover:bg-white/10 rounded-xl transition-colors ring-1 ring-white/5 mt-2"
                    onClick={onFocus}
                    title="Focus Selected"
                >
                    <Focus size={20} />
                </Button>
            )}
        </div>
    );
}

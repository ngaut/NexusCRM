import React from 'react';
import { useNavigate } from 'react-router-dom';
import { FileQuestion, ArrowLeft, Home } from 'lucide-react';
import { Button } from '../components/ui/Button';

export const NotFound: React.FC = () => {
    const navigate = useNavigate();

    return (
        <div className="min-h-screen bg-slate-50 flex items-center justify-center p-6">
            <div className="text-center max-w-md mx-auto">
                <div className="w-24 h-24 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-6">
                    <FileQuestion className="w-12 h-12 text-slate-400" />
                </div>

                <h1 className="text-4xl font-bold text-slate-900 mb-3">Page Not Found</h1>

                <p className="text-lg text-slate-600 mb-8">
                    The page you are looking for doesn't exist or has been moved.
                </p>

                <div className="flex items-center justify-center gap-4">
                    <Button
                        variant="secondary"
                        onClick={() => navigate(-1)}
                        icon={<ArrowLeft className="w-4 h-4" />}
                    >
                        Go Back
                    </Button>
                    <Button
                        variant="primary"
                        onClick={() => navigate('/')}
                        icon={<Home className="w-4 h-4" />}
                    >
                        Go Home
                    </Button>
                </div>
            </div>
        </div>
    );
};
